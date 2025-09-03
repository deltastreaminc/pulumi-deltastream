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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"

	ds "github.com/deltastreaminc/go-deltastream"
)

// DeltaStreamObject represents a DeltaStream relation (stream, changelog, table).
type DeltaStreamObject struct{}

func (o *DeltaStreamObject) Annotate(a infer.Annotator) {
	a.Describe(o, "DeltaStreamObject (relation) resource supporting STREAM, CHANGELOG, TABLE creation via SQL DDL")
}

// DeltaStreamObjectArgs defines user supplied inputs
// DeltaStreamObjectArgs defines user provided inputs for creating a relation.
type DeltaStreamObjectArgs struct {
	// Database containing the relation.
	Database string `pulumi:"database"`
	// Namespace (schema) for the relation.
	Namespace string `pulumi:"namespace"`
	// Store backing the relation.
	Store string `pulumi:"store"`
	// SQL DDL statement (CREATE STREAM/CHANGELOG/TABLE ...).
	SQL string `pulumi:"sql"`
	// Owning role (optional) overriding provider role.
	Owner *string `pulumi:"owner,optional"`
}

// DeltaStreamObjectState extends inputs with computed fields
// DeltaStreamObjectState extends inputs with computed fields.
type DeltaStreamObjectState struct {
	DeltaStreamObjectArgs
	// Object name extracted from plan.
	Name string `pulumi:"name"`
	// Path segments [database, namespace, name].
	Path []string `pulumi:"path"`
	// Fully qualified name "db"."schema"."name".
	FQN string `pulumi:"fqn"`
	// Normalized type (stream|changelog|table).
	Type string `pulumi:"type"`
	// Provisioning state.
	State string `pulumi:"state"`
	// Owner as returned from system (may differ from input Owner during update).
	OwnerOut *string `pulumi:"owner"`
	// Creation timestamp.
	CreatedAt string `pulumi:"createdAt"`
	// Last update timestamp.
	UpdatedAt string `pulumi:"updatedAt"`
}

func (s *DeltaStreamObjectState) Annotate(a infer.Annotator) {
	a.Describe(&s.Type, "Type of the relation (stream|changelog|table)")
	a.Describe(&s.State, "Provisioning state of the relation")
	a.Describe(&s.Path, "Path of the relation as [database, namespace, name]")
	a.Describe(&s.FQN, "Fully qualified name of the relation as database.namespace.name")
}

// Planning output structure (subset)
type plannerRelation struct {
	FQN        string `json:"fqn"`
	Type       string `json:"type"`
	DbName     string `json:"db_name"`
	SchemaName string `json:"schema_name"`
	Name       string `json:"name"`
	StoreName  string `json:"store_name"`
}
type plannerStatement struct {
	Ddl *plannerRelation `json:"ddl"`
}

type artifactDDL struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Command string `json:"command"`
	Summary string `json:"summary"`
	Path    string `json:"path"`
}

// Check validates user input by planning the statement with DESCRIBE instead of regex parsing.
func (DeltaStreamObject) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[DeltaStreamObjectArgs], error) {
	args, failures, err := infer.DefaultCheck[DeltaStreamObjectArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[DeltaStreamObjectArgs]{}, err
	}

	// Require core scoping fields before attempting planning
	if args.SQL == "" || args.Database == "" || args.Namespace == "" || args.Store == "" {
		return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
	}

	cfg := infer.GetConfig[Config](ctx)
	db, openErr := openDB(ctx, &cfg)
	if openErr != nil { // tolerate inability to connect during preview
		return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
	}
	defer db.Close()

	role := ptr.Deref(args.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, crErr := withOrgRole(ctx, db, org, role)
	if crErr != nil { // skip planning if role/org application fails
		return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
	}
	defer conn.Close()

	if err := setSQLContext(conn, args.Database, args.Namespace, args.Store); err != nil {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("failed to set context: %v", err)})
		return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
	}

	kind, plan, dErr := describeStatement(ctx2, conn, args.SQL)
	if dErr != nil {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("describe failed: %v", dErr)})
		return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
	}

	if _, kErr := normalizeKind(kind); kErr != nil {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: kErr.Error()})
		return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
	}
	if plan.Ddl == nil {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: "describe returned empty plan"})
		return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
	}

	if plan.Ddl.DbName != args.Database {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("database mismatch: statement targets %s", plan.Ddl.DbName)})
	}
	if plan.Ddl.SchemaName != args.Namespace {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("namespace mismatch: statement targets %s", plan.Ddl.SchemaName)})
	}
	if plan.Ddl.StoreName != args.Store {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("store mismatch: statement targets %s", plan.Ddl.StoreName)})
	}

	return infer.CheckResponse[DeltaStreamObjectArgs]{Inputs: args, Failures: failures}, nil
}

// Diff decides replace vs update
// Diff decides whether changes require replacement or can be updated in-place.
func (DeltaStreamObject) Diff(ctx context.Context, req infer.DiffRequest[DeltaStreamObjectArgs, DeltaStreamObjectState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Database != "" && req.State.Database != req.Inputs.Database {
		diff["database"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.State.Namespace != "" && req.State.Namespace != req.Inputs.Namespace {
		diff["namespace"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.State.Store != "" && req.State.Store != req.Inputs.Store {
		diff["store"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.State.SQL != "" && req.State.SQL != req.Inputs.SQL {
		diff["sql"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if (req.State.Owner == nil && req.Inputs.Owner != nil) || (req.State.Owner != nil && req.Inputs.Owner != nil && *req.State.Owner != *req.Inputs.Owner) {
		diff["owner"] = p.PropertyDiff{Kind: p.Update}
	}
	return infer.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}

// Create executes planning + creation + polling
// Create plans and executes creation of the relation, polling until ready.
func (DeltaStreamObject) Create(ctx context.Context, req infer.CreateRequest[DeltaStreamObjectArgs]) (infer.CreateResponse[DeltaStreamObjectState], error) {
	in := req.Inputs
	logger := p.GetLogger(ctx)
	logger.Debug(fmt.Sprintf("Planning object create in %s.%s store=%s", in.Database, in.Namespace, in.Store))

	if req.DryRun {
		now := time.Now().UTC().Format(time.RFC3339)
		st := DeltaStreamObjectState{DeltaStreamObjectArgs: in, CreatedAt: now, UpdatedAt: now}
		return infer.CreateResponse[DeltaStreamObjectState]{ID: provisionalID(in), Output: st}, nil
	}

	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(in.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, err
	}
	defer conn.Close()

	if err := setSQLContext(conn, in.Database, in.Namespace, in.Store); err != nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, fmt.Errorf("failed setting sql context: %w", err)
	}

	kind, plan, err := describeStatement(ctx, conn, in.SQL)
	if err != nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, err
	}
	if plan.Ddl == nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, fmt.Errorf("planning error: invalid object plan")
	}
	if plan.Ddl.DbName != in.Database {
		return infer.CreateResponse[DeltaStreamObjectState]{}, fmt.Errorf("planning error: database name mismatch, statement would create object in %s instead of %s", plan.Ddl.DbName, in.Database)
	}
	if plan.Ddl.SchemaName != in.Namespace {
		return infer.CreateResponse[DeltaStreamObjectState]{}, fmt.Errorf("planning error: namespace name mismatch, statement would create object in %s instead of %s", plan.Ddl.SchemaName, in.Namespace)
	}
	if plan.Ddl.StoreName != in.Store {
		return infer.CreateResponse[DeltaStreamObjectState]{}, fmt.Errorf("planning error: store name mismatch, statement would use store %s instead of %s", plan.Ddl.StoreName, in.Store)
	}
	typNormalized, err := normalizeKind(kind)
	if err != nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, err
	}

	art, err := executeCreate(ctx, conn, in.SQL)
	if err != nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, fmt.Errorf("failed to create relation: %w", err)
	}

	var pathArr []string
	if err = json.Unmarshal([]byte(art.Path), &pathArr); err != nil {
		return infer.CreateResponse[DeltaStreamObjectState]{}, fmt.Errorf("failed to parse object path: %w", err)
	}

	fqn := getFQN(pathArr)
	row, err := waitForRelationReady(ctx, conn, pathArr)
	if err != nil {
		_ = dropRelation(ctx, conn, fqn)
		return infer.CreateResponse[DeltaStreamObjectState]{}, err
	}
	id := getFQN(pathArr)

	ownerOut := row.Owner
	state := DeltaStreamObjectState{DeltaStreamObjectArgs: in, Name: art.Name, Path: pathArr, FQN: fqn, Type: typNormalized, State: row.State, OwnerOut: &ownerOut, CreatedAt: row.CreatedAt.Format(time.RFC3339), UpdatedAt: row.UpdatedAt.Format(time.RFC3339)}
	logger.Info(fmt.Sprintf("Object created: %s (%s)", art.Name, state.Type))
	return infer.CreateResponse[DeltaStreamObjectState]{ID: id, Output: state}, nil
}

// Read refreshes state
// Read refreshes state from the system catalogs.
func (DeltaStreamObject) Read(ctx context.Context, req infer.ReadRequest[DeltaStreamObjectArgs, DeltaStreamObjectState]) (infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState], error) {
	if req.ID == "" {
		return infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState]{}, nil
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState]{}, err
	}
	defer conn.Close()

	row, err := lookupRelation(ctx, conn, req.State.Path)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidRelation {
			return infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState]{}, nil
		}
		if errors.Is(err, sql.ErrNoRows) {
			return infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState]{}, nil
		}
		return infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState]{}, err
	}
	typ := strings.ToLower(row.Type)
	ownerOut := row.Owner
	st := req.State
	st.Type = typ
	st.State = row.State
	st.OwnerOut = &ownerOut
	st.CreatedAt = row.CreatedAt.Format(time.RFC3339)
	st.UpdatedAt = row.UpdatedAt.Format(time.RFC3339)
	st.Name = row.Name
	return infer.ReadResponse[DeltaStreamObjectArgs, DeltaStreamObjectState]{ID: req.ID, Inputs: st.DeltaStreamObjectArgs, State: st}, nil
}

// Update only supports owner change (future)
// Update changes supported mutable fields (currently only owner).
func (DeltaStreamObject) Update(ctx context.Context, req infer.UpdateRequest[DeltaStreamObjectArgs, DeltaStreamObjectState]) (infer.UpdateResponse[DeltaStreamObjectState], error) {
	st := req.State
	// Only owner changes are supported
	if (st.Owner == nil && req.Inputs.Owner == nil) || (st.Owner != nil && req.Inputs.Owner != nil && *st.Owner == *req.Inputs.Owner) {
		return infer.UpdateResponse[DeltaStreamObjectState]{}, nil
	}
	if req.DryRun {
		st.Owner = req.Inputs.Owner
		return infer.UpdateResponse[DeltaStreamObjectState]{Output: st}, nil
	}
	if len(st.Path) == 0 {
		return infer.UpdateResponse[DeltaStreamObjectState]{}, fmt.Errorf("missing path for update")
	}
	if req.Inputs.Owner == nil {
		return infer.UpdateResponse[DeltaStreamObjectState]{}, fmt.Errorf("owner update requires non-nil owner")
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.UpdateResponse[DeltaStreamObjectState]{}, err
	}
	defer db.Close()
	// Use provider role (not the new owner yet) to perform alteration; assume it has privileges
	role := ptr.Deref(st.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.UpdateResponse[DeltaStreamObjectState]{}, err
	}
	defer conn.Close()
	typ := strings.ToUpper(st.Type)
	stmt := fmt.Sprintf("ALTER %s %s OWNER TO %s;", typ, getFQN(st.Path), *req.Inputs.Owner)
	if _, err := conn.ExecContext(ctx2, stmt); err != nil {
		return infer.UpdateResponse[DeltaStreamObjectState]{}, fmt.Errorf("failed altering owner: %w", err)
	}
	// Re-query to refresh owner & timestamps
	row, err := lookupRelation(ctx2, conn, st.Path)
	if err != nil {
		return infer.UpdateResponse[DeltaStreamObjectState]{}, err
	}
	newOwner := row.Owner
	st.Owner = req.Inputs.Owner
	st.OwnerOut = &newOwner
	st.UpdatedAt = row.UpdatedAt.Format(time.RFC3339)
	return infer.UpdateResponse[DeltaStreamObjectState]{Output: st}, nil
}

// Delete drops the relation
// Delete drops the relation and waits for its removal.
func (DeltaStreamObject) Delete(ctx context.Context, req infer.DeleteRequest[DeltaStreamObjectState]) (infer.DeleteResponse, error) {
	if req.ID == "" {
		return infer.DeleteResponse{}, nil
	}
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
	_ = dropRelation(ctx, conn, req.State.FQN)
	_ = waitForRelationGone(ctx, conn, req.State.Path, time.Minute)
	return infer.DeleteResponse{}, nil
}

// WireDependencies declares input-output relationships for Pulumi's graph.
func (DeltaStreamObject) WireDependencies(f infer.FieldSelector, args *DeltaStreamObjectArgs, state *DeltaStreamObjectState) {
	f.OutputField(&state.Name).DependsOn(f.InputField(&args.SQL))
	f.OutputField(&state.Path).DependsOn(f.InputField(&args.SQL))
	f.OutputField(&state.Type).DependsOn(f.InputField(&args.SQL))
	f.OutputField(&state.State).DependsOn(f.InputField(&args.SQL))
}

func describeStatement(ctx context.Context, conn *sql.Conn, sqlText string) (kind string, plan plannerStatement, err error) {
	q := fmt.Sprintf("DESCRIBE %s", sqlText)
	// Acquire both columns in one scan; driver must support it.
	row := conn.QueryRowContext(ctx, q)
	var planJSON string
	if err := row.Scan(&kind, &planJSON); err != nil {
		return "", plan, fmt.Errorf("describe failed: %w", err)
	}
	if !strings.HasPrefix(strings.ToUpper(kind), "CREATE_") {
		return "", plan, fmt.Errorf("invalid relation kind: %s", kind)
	}
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return "", plan, fmt.Errorf("failed to parse plan json: %w", err)
	}
	return kind, plan, nil
}

func normalizeKind(kind string) (string, error) {
	switch strings.ToUpper(kind) {
	case "CREATE_STREAM":
		return "stream", nil
	case "CREATE_CHANGELOG":
		return "changelog", nil
	case "CREATE_TABLE":
		return "table", nil
	default:
		return "", fmt.Errorf("invalid relation type: %s", kind)
	}
}

func executeCreate(ctx context.Context, conn *sql.Conn, sqlText string) (artifactDDL, error) {
	row := conn.QueryRowContext(ctx, sqlText)
	var art artifactDDL
	if err := row.Scan(&art.Type, &art.Name, &art.Command, &art.Summary, &art.Path); err != nil {
		return art, err
	}
	return art, nil
}

type relationRow struct {
	Name, Type, Owner, State string
	CreatedAt, UpdatedAt     time.Time
}

func lookupRelation(ctx context.Context, conn *sql.Conn, path []string) (relationRow, error) {
	q := fmt.Sprintf(`SELECT name, relation_type, "owner", "state", created_at, updated_at 
		FROM deltastream.sys."relations" WHERE database_name = %s AND schema_name = %s AND name = %s;`, quoteString(path[0]), quoteString(path[1]), quoteString(path[2]))
	row := conn.QueryRowContext(ctx, q)
	var r relationRow
	if err := row.Scan(&r.Name, &r.Type, &r.Owner, &r.State, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return r, err
	}
	return r, nil
}

func waitForRelationReady(ctx context.Context, conn *sql.Conn, path []string) (relationRow, error) {
	deadline := time.Now().Add(5 * time.Minute)
	var lastErr error

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		row, err := lookupRelation(ctx, conn, path)
		if err == nil {
			return row, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			lastErr = err
		}
		if time.Now().After(deadline) {
			if lastErr != nil {
				return relationRow{}, fmt.Errorf("timeout waiting relation %s: %w", strings.Join(path, "."), lastErr)
			}
			return relationRow{}, fmt.Errorf("timeout waiting relation %s", strings.Join(path, "."))
		}
		select {
		case <-t.C:
		case <-ctx.Done():
			return relationRow{}, ctx.Err()
		}
	}
}

func dropRelation(ctx context.Context, conn *sql.Conn, fqn string) error {
	stmt := fmt.Sprintf("DROP RELATION %s;", fqn)
	_, err := conn.ExecContext(ctx, stmt)
	return err
}

func waitForRelationGone(ctx context.Context, conn *sql.Conn, path []string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		q := fmt.Sprintf(`SELECT 1 FROM deltastream.sys."relations" WHERE database_name=%s AND schema_name=%s AND name=%s;`, quoteString(path[0]), quoteString(path[1]), quoteString(path[2]))
		row := conn.QueryRowContext(ctx, q)
		var one int
		err := row.Scan(&one)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting deletion of %s", strings.Join(path, "."))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

func setSQLContext(conn *sql.Conn, database, schema, store string) error {
	return conn.Raw(func(dc interface{}) error {
		if c, ok := dc.(*ds.Conn); ok {
			rsctx := c.GetContext()
			if database != "" {
				rsctx.DatabaseName = &database
			}
			if schema != "" {
				rsctx.SchemaName = &schema
			}
			if store != "" {
				rsctx.StoreName = &store
			}
			c.SetContext(rsctx)
		}
		return nil
	})
}

func provisionalID(in DeltaStreamObjectArgs) string {
	return fmt.Sprintf("%s/%s/%s", in.Database, in.Namespace, in.Store)
}

type GetObjectArgs struct {
	Database  string `pulumi:"database"`
	Namespace string `pulumi:"namespace"`
	Name      string `pulumi:"name"`
}
type GetObjectResult struct {
	Database  string `pulumi:"database"`
	Namespace string `pulumi:"namespace"`
	Name      string `pulumi:"name"`
	FQN       string `pulumi:"fqn"`
	Type      string `pulumi:"type"`
	State     string `pulumi:"state"`
	Owner     string `pulumi:"owner"`
	CreatedAt string `pulumi:"createdAt"`
	UpdatedAt string `pulumi:"updatedAt"`
}

type GetObject struct{}

func (GetObject) Invoke(ctx context.Context, req infer.FunctionRequest[GetObjectArgs]) (infer.FunctionResponse[GetObjectResult], error) {
	args := req.Input
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetObjectResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetObjectResult]{}, err
	}
	defer conn.Close()
	q := fmt.Sprintf(`SELECT name, fqn, relation_type, "owner", "state", created_at, updated_at FROM deltastream.sys."relations" WHERE database_name = %s AND schema_name = %s AND name = %s;`, quoteString(args.Database), quoteString(args.Namespace), quoteString(args.Name))
	row := conn.QueryRowContext(ctx, q)
	var name, fqn, typ, owner, state string
	var created, updated time.Time
	if err := row.Scan(&name, &fqn, &typ, &owner, &state, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return infer.FunctionResponse[GetObjectResult]{}, fmt.Errorf("object %s.%s.%s not found", args.Database, args.Namespace, args.Name)
		}
		return infer.FunctionResponse[GetObjectResult]{}, err
	}
	res := GetObjectResult{Database: args.Database, Namespace: args.Namespace, Name: name, FQN: fqn, Type: strings.ToLower(typ), State: state, Owner: owner, CreatedAt: created.Format(time.RFC3339), UpdatedAt: updated.Format(time.RFC3339)}
	return infer.FunctionResponse[GetObjectResult]{Output: res}, nil
}

type GetObjectsArgs struct {
	Database  string `pulumi:"database"`
	Namespace string `pulumi:"namespace"`
}
type GetObjectsResult struct {
	Objects []GetObjectResult `pulumi:"objects"`
}
type GetObjects struct{}

func (GetObjects) Invoke(ctx context.Context, req infer.FunctionRequest[GetObjectsArgs]) (infer.FunctionResponse[GetObjectsResult], error) {
	args := req.Input
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetObjectsResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetObjectsResult]{}, err
	}
	defer conn.Close()
	q := fmt.Sprintf(`SELECT name, fqn, relation_type, "owner", "state", created_at, updated_at FROM deltastream.sys."relations" WHERE database_name = %s AND schema_name = %s;`, quoteString(args.Database), quoteString(args.Namespace))
	rows, err := conn.QueryContext(ctx, q)
	if err != nil {
		return infer.FunctionResponse[GetObjectsResult]{}, err
	}
	defer rows.Close()
	var list []GetObjectResult
	for rows.Next() {
		var name, fqn, typ, owner, state string
		var created, updated time.Time
		if err := rows.Scan(&name, &fqn, &typ, &owner, &state, &created, &updated); err != nil {
			return infer.FunctionResponse[GetObjectsResult]{}, err
		}
		list = append(list, GetObjectResult{Database: args.Database, Namespace: args.Namespace, Name: name, FQN: fqn, Type: strings.ToLower(typ), State: state, Owner: owner, CreatedAt: created.Format(time.RFC3339), UpdatedAt: updated.Format(time.RFC3339)})
	}
	if err := rows.Err(); err != nil {
		return infer.FunctionResponse[GetObjectsResult]{}, err
	}
	return infer.FunctionResponse[GetObjectsResult]{Output: GetObjectsResult{Objects: list}}, nil
}

func getFQN(path []string) string {
	return fmt.Sprintf(`"%s"."%s"."%s"`, path[0], path[1], path[2])
}
