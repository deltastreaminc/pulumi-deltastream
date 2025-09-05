// Copyright 2025, DeltaStream Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"

	ds "github.com/deltastreaminc/go-deltastream"
)

// Namespace resource represents a schema within a database.
type Namespace struct{}

func (n *Namespace) Annotate(a infer.Annotator) {
	a.Describe(&n, "Namespace resource providing logical grouping within a database for streams and other objects")
}

// NamespaceArgs are inputs to create a namespace.
type NamespaceArgs struct {
	Database string  `pulumi:"database"`
	Name     string  `pulumi:"name"`
	Owner    *string `pulumi:"owner,optional"`
}

func (a *NamespaceArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Database, "Name of the database containing the namespace")
	an.Describe(&a.Name, "Name of the namespace")
	an.Describe(&a.Owner, "Optional owning role. When set, statements execute as this role during create.")
}

// NamespaceState persists namespace data.
type NamespaceState struct {
	NamespaceArgs
	CreatedAt string `pulumi:"createdAt"`
}

func (s *NamespaceState) Annotate(a infer.Annotator) {
	a.Describe(&s.CreatedAt, "Creation timestamp of the namespace")
}

// Create a namespace.
func (Namespace) Create(ctx context.Context, req infer.CreateRequest[NamespaceArgs]) (infer.CreateResponse[NamespaceState], error) {
	in := req.Inputs
	logger := p.GetLogger(ctx)
	logger.Debug(fmt.Sprintf("Creating namespace %s.%s", in.Database, in.Name))

	if req.DryRun {
		state := NamespaceState{NamespaceArgs: in, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
		return infer.CreateResponse[NamespaceState]{ID: fmt.Sprintf("%s/%s", in.Database, in.Name), Output: state}, nil
	}

	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.CreateResponse[NamespaceState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(in.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")

	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.CreateResponse[NamespaceState]{}, err
	}
	defer conn.Close()

	stmt := fmt.Sprintf("CREATE SCHEMA %s IN DATABASE %s;", quoteIdent(in.Name), quoteIdent(in.Database))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return infer.CreateResponse[NamespaceState]{}, fmt.Errorf("failed to create namespace: %w", err)
	}

	owner, createdAt, err := lookupNamespace(ctx, conn, in.Database, in.Name)
	if err != nil {
		// best effort cleanup
		_, _ = conn.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA %s.%s;", quoteIdent(in.Database), quoteIdent(in.Name)))
		return infer.CreateResponse[NamespaceState]{}, fmt.Errorf("failed to verify namespace: %w", err)
	}

	state := NamespaceState{NamespaceArgs: NamespaceArgs{Database: in.Database, Name: in.Name, Owner: &owner}, CreatedAt: createdAt.Format(time.RFC3339)}
	logger.Info(fmt.Sprintf("Namespace created: %s.%s", in.Database, in.Name))
	return infer.CreateResponse[NamespaceState]{ID: fmt.Sprintf("%s/%s", in.Database, in.Name), Output: state}, nil
}

// Read retrieves namespace state.
func (Namespace) Read(ctx context.Context, req infer.ReadRequest[NamespaceArgs, NamespaceState]) (infer.ReadResponse[NamespaceArgs, NamespaceState], error) {
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.ReadResponse[NamespaceArgs, NamespaceState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.ReadResponse[NamespaceArgs, NamespaceState]{}, err
	}
	defer conn.Close()

	dbName := req.State.Database
	nsName := req.State.Name
	owner, createdAt, err := lookupNamespace(ctx, conn, dbName, nsName)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidSchema {
			return infer.ReadResponse[NamespaceArgs, NamespaceState]{}, nil
		}
		return infer.ReadResponse[NamespaceArgs, NamespaceState]{}, err
	}
	state := NamespaceState{NamespaceArgs: NamespaceArgs{Database: dbName, Name: nsName, Owner: &owner}, CreatedAt: createdAt.Format(time.RFC3339)}
	return infer.ReadResponse[NamespaceArgs, NamespaceState]{ID: fmt.Sprintf("%s/%s", dbName, nsName), Inputs: state.NamespaceArgs, State: state}, nil
}

// Update not supported.
func (Namespace) Update(ctx context.Context, req infer.UpdateRequest[NamespaceArgs, NamespaceState]) (infer.UpdateResponse[NamespaceState], error) {
	return infer.UpdateResponse[NamespaceState]{}, fmt.Errorf("namespace updates not supported")
}

// Delete namespace.
func (Namespace) Delete(ctx context.Context, req infer.DeleteRequest[NamespaceState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	defer db.Close()

	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	defer conn.Close()

	stmt := fmt.Sprintf("DROP SCHEMA %s.%s;", quoteIdent(req.State.Database), quoteIdent(req.State.Name))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		var sqlErr ds.ErrSQLError
		if !errors.As(err, &sqlErr) || (sqlErr.SQLCode != ds.SqlStateInvalidDatabase && sqlErr.SQLCode != ds.SqlStateInvalidSchema) {
			return infer.DeleteResponse{}, fmt.Errorf("failed to delete namespace: %w", err)
		}
	}
	return infer.DeleteResponse{}, nil
}

// Diff for namespace.
func (Namespace) Diff(ctx context.Context, req infer.DiffRequest[NamespaceArgs, NamespaceState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Database != req.Inputs.Database {
		diff["database"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.State.Name != req.Inputs.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if (req.State.Owner == nil && req.Inputs.Owner != nil) || (req.State.Owner != nil && req.Inputs.Owner != nil && *req.State.Owner != *req.Inputs.Owner) {
		diff["owner"] = p.PropertyDiff{Kind: p.Update}
	}
	return infer.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}

// Check validates namespace inputs.
func (Namespace) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[NamespaceArgs], error) {
	args, failures, err := infer.DefaultCheck[NamespaceArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[NamespaceArgs]{}, err
	}
	return infer.CheckResponse[NamespaceArgs]{Inputs: args, Failures: failures}, nil
}

// WireDependencies sets output dependencies.
func (Namespace) WireDependencies(f infer.FieldSelector, args *NamespaceArgs, state *NamespaceState) {
	f.OutputField(&state.Database).DependsOn(f.InputField(&args.Database))
	f.OutputField(&state.Name).DependsOn(f.InputField(&args.Name))
	f.OutputField(&state.Owner).DependsOn(f.InputField(&args.Owner))
}

// lookupNamespace queries owner and created_at for namespace.
func lookupNamespace(ctx context.Context, conn *sql.Conn, dbName, nsName string) (owner string, createdAt time.Time, err error) {
	q := fmt.Sprintf(`SELECT "owner", created_at FROM deltastream.sys."schemas" WHERE database_name = %s AND name = %s;`, quoteString(dbName), quoteString(nsName))
	logger := p.GetLogger(ctx)
	logger.Debug(q)
	row := conn.QueryRowContext(ctx, q)
	if err := row.Err(); err != nil {
		return "", time.Time{}, err
	}
	if err := row.Scan(&owner, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, ds.ErrSQLError{SQLCode: ds.SqlStateInvalidSchema}
		}
		return "", time.Time{}, err
	}
	return owner, createdAt, nil
}
