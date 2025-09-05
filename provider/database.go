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

// Database is the controller for the DeltaStream database resource.
type Database struct{}

func (d *Database) Annotate(a infer.Annotator) {
	a.Describe(&d, "A DeltaStream database resource that provides namespacing for streams, changelogs, and other objects")
}

// DatabaseArgs are the inputs to the database resource's constructor.
type DatabaseArgs struct {
	// Name of the database
	Name string `pulumi:"name"`
	// Owning role; overrides provider role for creation (optional)
	Owner *string `pulumi:"owner,optional"`
}

func (d *DatabaseArgs) Annotate(a infer.Annotator) {
	a.Describe(&d.Name, "The name of the database to create. If the name is case sensitive, wrap it in quotes.")
	a.Describe(&d.Owner, "Optional owning role. When set, statements execute as this role during create.")
}

// DatabaseState is what's persisted in state.
type DatabaseState struct {
	// Embed the args to include them in outputs
	DatabaseArgs
	// Timestamp when the database was created
	CreatedAt string `pulumi:"createdAt"`
}

func (d *DatabaseState) Annotate(a infer.Annotator) {
	a.Describe(&d.Owner, "The owner of the database")
	a.Describe(&d.CreatedAt, "The timestamp when the database was created")
}

// Create creates a new instance of the database resource.
func (Database) Create(ctx context.Context, req infer.CreateRequest[DatabaseArgs]) (infer.CreateResponse[DatabaseState], error) {
	input := req.Inputs
	logger := p.GetLogger(ctx)
	logger.Debug(fmt.Sprintf("Creating database %s", input.Name))

	if req.DryRun {
		state := DatabaseState{DatabaseArgs: input, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
		return infer.CreateResponse[DatabaseState]{ID: input.Name, Output: state}, nil
	}

	// Open connection using provider config
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.CreateResponse[DatabaseState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(input.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")

	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.CreateResponse[DatabaseState]{}, err
	}
	defer conn.Close()

	// CREATE DATABASE using identifier quoting helper (quoteIdent already adds quotes)
	stmt := fmt.Sprintf("CREATE DATABASE %s;", quoteIdent(input.Name))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return infer.CreateResponse[DatabaseState]{}, fmt.Errorf("failed to create database: %w", err)
	}

	// Lookup owner and created_at
	owner, createdAt, err := lookupDatabase(ctx, conn, input.Name)
	if err != nil {
		// best effort rollback
		_, _ = conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s;", quoteIdent(input.Name)))
		return infer.CreateResponse[DatabaseState]{}, fmt.Errorf("failed to verify database creation: %w", err)
	}

	state := DatabaseState{
		DatabaseArgs: DatabaseArgs{
			Name:  input.Name,
			Owner: &owner,
		},
		CreatedAt: createdAt.Format(time.RFC3339),
	}

	logger.Info(fmt.Sprintf("Database created successfully: %s", input.Name))
	return infer.CreateResponse[DatabaseState]{ID: input.Name, Output: state}, nil
}

// Read fetches the resource's state from the provider.
func (Database) Read(
	ctx context.Context,
	req infer.ReadRequest[DatabaseArgs, DatabaseState],
) (infer.ReadResponse[DatabaseArgs, DatabaseState], error) {
	logger := p.GetLogger(ctx)
	logger.Debug(fmt.Sprintf("Reading database with ID: %s", req.ID))

	// Open connection using provider config
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.ReadResponse[DatabaseArgs, DatabaseState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")

	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.ReadResponse[DatabaseArgs, DatabaseState]{}, err
	}
	defer conn.Close()

	owner, createdAt, err := lookupDatabase(ctx, conn, req.ID)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidDatabase {
			return infer.ReadResponse[DatabaseArgs, DatabaseState]{}, nil
		}
		return infer.ReadResponse[DatabaseArgs, DatabaseState]{}, fmt.Errorf("failed to read database: %w", err)
	}

	state := DatabaseState{
		DatabaseArgs: DatabaseArgs{
			Name:  req.ID,
			Owner: &owner,
		},
		CreatedAt: createdAt.Format(time.RFC3339),
	}

	return infer.ReadResponse[DatabaseArgs, DatabaseState]{ID: req.ID, Inputs: state.DatabaseArgs, State: state}, nil
}

// Update updates an existing database.
func (Database) Update(ctx context.Context, req infer.UpdateRequest[DatabaseArgs, DatabaseState]) (infer.UpdateResponse[DatabaseState], error) {
	logger := p.GetLogger(ctx)
	logger.Debug(fmt.Sprintf("Updating database with ID: %s", req.ID))

	if req.DryRun {
		return infer.UpdateResponse[DatabaseState]{}, nil
	}
	return infer.UpdateResponse[DatabaseState]{}, fmt.Errorf("database updates not supported")
}

// Delete deletes an existing database.
func (Database) Delete(ctx context.Context, req infer.DeleteRequest[DatabaseState]) (infer.DeleteResponse, error) {
	logger := p.GetLogger(ctx)
	logger.Debug(fmt.Sprintf("Deleting database with ID: %s", req.ID))

	// Open connection using provider config
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

	if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s;", quoteIdent(req.ID))); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete database: %w", err)
	}

	logger.Info(fmt.Sprintf("Database deleted successfully: %s", req.ID))
	return infer.DeleteResponse{}, nil
}

// Diff computes the difference between the current and desired state.
func (Database) Diff(ctx context.Context, req infer.DiffRequest[DatabaseArgs, DatabaseState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Name != req.Inputs.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if (req.State.Owner == nil && req.Inputs.Owner != nil) || (req.State.Owner != nil && req.Inputs.Owner != nil && *req.State.Owner != *req.Inputs.Owner) {
		diff["owner"] = p.PropertyDiff{Kind: p.Update}
	}
	return infer.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}

// Check validates the input parameters.
func (Database) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[DatabaseArgs], error) {
	args, failures, err := infer.DefaultCheck[DatabaseArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[DatabaseArgs]{}, err
	}
	return infer.CheckResponse[DatabaseArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// WireDependencies defines the dependencies between inputs and outputs.
func (Database) WireDependencies(f infer.FieldSelector, args *DatabaseArgs, state *DatabaseState) {
	f.OutputField(&state.Name).DependsOn(f.InputField(&args.Name))
	f.OutputField(&state.Owner).DependsOn(f.InputField(&args.Owner))
}

// lookupDatabase queries owner and created_at for a database
func lookupDatabase(ctx context.Context, conn *sql.Conn, name string) (owner string, createdAt time.Time, err error) {
	q := fmt.Sprintf(`SELECT "owner", created_at FROM deltastream.sys."databases" WHERE name = %s;`, quoteString(name))
	logger := p.GetLogger(ctx)
	logger.Debug(q)
	row := conn.QueryRowContext(ctx, q)
	if err := row.Err(); err != nil {
		return "", time.Time{}, err
	}
	if err := row.Scan(&owner, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, ds.ErrSQLError{SQLCode: ds.SqlStateInvalidDatabase}
		}
		return "", time.Time{}, err
	}
	return owner, createdAt, nil
}
