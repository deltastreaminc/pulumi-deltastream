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
	"regexp"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"

	ds "github.com/deltastreaminc/go-deltastream"
)

// Query resource implements continuous INSERT INTO ... SELECT ... queries.
type Query struct{}

func (q *Query) Annotate(a infer.Annotator) {
	a.Describe(q, "Continuous query resource (INSERT INTO ... SELECT ...) streaming data from source relations into a sink relation.")
}

// QueryArgs define the desired continuous query expressed as an INSERT INTO
// ... SELECT ... statement along with explicit enumeration of source and sink
// relations. The SQL must be an INSERT INTO form; validation runs a DESCRIBE
// to confirm semantic alignment.
type QueryArgs struct {
	SourceRelationFqns []string `pulumi:"sourceRelationFqns"`
	SinkRelationFqn    string   `pulumi:"sinkRelationFqn"`
	SQL                string   `pulumi:"sql"`
	Owner              *string  `pulumi:"owner,optional"`
}

// QueryState captures runtime attributes of a continuous query after creation.
// QueryID is the system identifier used for subsequent lifecycle operations.
type QueryState struct {
	QueryArgs
	QueryID      string  `pulumi:"queryId"`
	QueryName    *string `pulumi:"queryName,optional"`
	QueryVersion *int64  `pulumi:"queryVersion,optional"`
	State        string  `pulumi:"state"`
	CreatedAt    string  `pulumi:"createdAt"`
	UpdatedAt    string  `pulumi:"updatedAt"`
	OwnerOut     *string `pulumi:"owner"`
}

func (s *QueryState) Annotate(a infer.Annotator) {
	a.Describe(&s.State, "Lifecycle state of the query (starting|running|terminate_requested|terminated|errored)")
	a.Describe(&s.QueryID, "System-generated query identifier")
}

// describeQueryPlan structures (subset) matching Terraform provider logic
type queryRelationPlan struct {
	Fqn        string `json:"fqn"`
	Type       string `json:"type"`
	DbName     string `json:"db_name"`
	SchemaName string `json:"schema_name"`
	Name       string `json:"name"`
	StoreName  string `json:"store_name"`
}
type queryStatementPlan struct {
	Ddl     *queryRelationPlan  `json:"ddl,omitempty"`
	Sink    *queryRelationPlan  `json:"sink,omitempty"`
	Sources []queryRelationPlan `json:"sources,omitempty"`
}

type queryArtifactDDL struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Command string `json:"command"`
	Summary string `json:"summary"`
	Path    string `json:"path"`
}

var insertIntoRe = regexp.MustCompile(`(?is)^\s*insert\s+into\b`)

// Check validates the SQL by running DESCRIBE and verifying plan alignment.
func (Query) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[QueryArgs], error) {
	args, failures, err := infer.DefaultCheck[QueryArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[QueryArgs]{}, err
	}
	if args.SQL == "" || args.SinkRelationFqn == "" || len(args.SourceRelationFqns) == 0 {
		return infer.CheckResponse[QueryArgs]{Inputs: args, Failures: failures}, nil
	}
	cfg := infer.GetConfig[Config](ctx)
	db, oerr := openDB(ctx, &cfg)
	if oerr != nil { // tolerate missing connection in preview
		return infer.CheckResponse[QueryArgs]{Inputs: args, Failures: failures}, nil
	}
	defer db.Close()
	role := ptr.Deref(args.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, cerr := withOrgRole(ctx, db, org, role)
	if cerr != nil {
		return infer.CheckResponse[QueryArgs]{Inputs: args, Failures: failures}, nil
	}
	defer conn.Close()

	kind, plan, derr := describeQuery(ctx2, conn, args.SQL)
	if derr != nil {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("describe failed: %v", derr)})
		return infer.CheckResponse[QueryArgs]{Inputs: args, Failures: failures}, nil
	}
	if kind != "INSERT_INTO" && !insertIntoRe.MatchString(args.SQL) { // fallback regex
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: "only INSERT INTO ... SELECT ... queries are supported"})
	}
	if plan.Sink == nil {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: "describe returned no sink relation"})
	} else if strings.TrimSpace(args.SinkRelationFqn) != getFQN([]string{plan.Sink.DbName, plan.Sink.SchemaName, plan.Sink.Name}) {
		failures = append(failures, p.CheckFailure{Property: "sinkRelationFqn", Reason: fmt.Sprintf("sink mismatch: %s", plan.Sink.Fqn)})
	}
	// Ensure every planned source is listed
	for _, src := range plan.Sources {
		found := false
		for _, declared := range args.SourceRelationFqns {
			if strings.TrimSpace(declared) == getFQN([]string{src.DbName, src.SchemaName, src.Name}) {
				found = true
				break
			}
		}
		if !found {
			failures = append(failures, p.CheckFailure{Property: "sourceRelationFqns", Reason: fmt.Sprintf("missing source %s", src.Fqn)})
		}
	}
	return infer.CheckResponse[QueryArgs]{Inputs: args, Failures: failures}, nil
}

// Diff: all fields replace except owner (update)
func (Query) Diff(ctx context.Context, req infer.DiffRequest[QueryArgs, QueryState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.SQL != "" && req.State.SQL != req.Inputs.SQL {
		diff["sql"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.State.SinkRelationFqn != "" && req.State.SinkRelationFqn != req.Inputs.SinkRelationFqn {
		diff["sinkRelationFqn"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if len(req.State.SourceRelationFqns) > 0 && !stringSlicesEqual(req.State.SourceRelationFqns, req.Inputs.SourceRelationFqns) {
		diff["sourceRelationFqns"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if (req.State.Owner == nil && req.Inputs.Owner != nil) || (req.State.Owner != nil && req.Inputs.Owner != nil && *req.State.Owner != *req.Inputs.Owner) {
		diff["owner"] = p.PropertyDiff{Kind: p.Update}
	}
	return infer.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}

// stringSlicesEqual returns true if slices have identical length and element order.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Create launches the query and waits until running
// Create validates the query SQL, executes it, and waits until the query reaches
// the running state before returning. On preview it returns a provisional ID.
func (Query) Create(ctx context.Context, req infer.CreateRequest[QueryArgs]) (infer.CreateResponse[QueryState], error) {
	in := req.Inputs
	logger := p.GetLogger(ctx)
	if req.DryRun {
		now := time.Now().UTC().Format(time.RFC3339)
		st := QueryState{QueryArgs: in, CreatedAt: now, UpdatedAt: now, State: "starting"}
		return infer.CreateResponse[QueryState]{ID: provisionalQueryID(&in), Output: st}, nil
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.CreateResponse[QueryState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(in.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.CreateResponse[QueryState]{}, err
	}
	defer conn.Close()

	kind, plan, derr := describeQuery(ctx2, conn, in.SQL)
	if derr != nil {
		return infer.CreateResponse[QueryState]{}, derr
	}
	if kind != "INSERT_INTO" && !insertIntoRe.MatchString(in.SQL) {
		return infer.CreateResponse[QueryState]{}, fmt.Errorf("unsupported query kind: %s", kind)
	}
	if plan.Sink == nil {
		return infer.CreateResponse[QueryState]{}, fmt.Errorf("planning error: missing sink relation")
	}
	if strings.TrimSpace(in.SinkRelationFqn) != getFQN([]string{plan.Sink.DbName, plan.Sink.SchemaName, plan.Sink.Name}) {
		return infer.CreateResponse[QueryState]{}, fmt.Errorf("sink relation mismatch %s != %s", org+"."+in.SinkRelationFqn, plan.Sink.Fqn)
	}
	for _, src := range plan.Sources {
		found := false
		for _, declared := range in.SourceRelationFqns {
			if strings.TrimSpace(declared) == getFQN([]string{src.DbName, src.SchemaName, src.Name}) {
				found = true
				break
			}
		}
		if !found {
			return infer.CreateResponse[QueryState]{}, fmt.Errorf("source relation %s not declared", src.Fqn)
		}
	}
	art, aerr := executeQuerySQL(ctx2, conn, in.SQL)
	if aerr != nil {
		return infer.CreateResponse[QueryState]{}, aerr
	}
	qid := art.Name
	// Poll for running state
	var qrow queryRow
	if perr := waitForQueryRunning(ctx2, conn, qid, time.Minute*10); perr != nil { // attempt cleanup
		_ = ensureQueryLookup(ctx2, conn, qid) // best-effort
		return infer.CreateResponse[QueryState]{}, perr
	}
	qrow, err = lookupQuery(ctx2, conn, qid)
	if err != nil {
		return infer.CreateResponse[QueryState]{}, err
	}
	ownerOut := qrow.Owner
	st := QueryState{QueryArgs: in, QueryID: qid, QueryName: qrow.Name, QueryVersion: qrow.Version, State: qrow.State, CreatedAt: qrow.CreatedAt.Format(time.RFC3339), UpdatedAt: qrow.UpdatedAt.Format(time.RFC3339), OwnerOut: &ownerOut}
	logger.Info(fmt.Sprintf("Query created: %s", qid))
	return infer.CreateResponse[QueryState]{ID: qid, Output: st}, nil
}

// Read refreshes state
// Read refreshes the status of the query and returns nil response if the query
// has been removed server-side.
func (Query) Read(ctx context.Context, req infer.ReadRequest[QueryArgs, QueryState]) (infer.ReadResponse[QueryArgs, QueryState], error) {
	if req.ID == "" {
		return infer.ReadResponse[QueryArgs, QueryState]{}, nil
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.ReadResponse[QueryArgs, QueryState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.ReadResponse[QueryArgs, QueryState]{}, err
	}
	defer conn.Close()
	qrow, err := lookupQuery(ctx2, conn, req.ID)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidQuery {
			return infer.ReadResponse[QueryArgs, QueryState]{}, nil
		}
		if errors.Is(err, sql.ErrNoRows) {
			return infer.ReadResponse[QueryArgs, QueryState]{}, nil
		}
		return infer.ReadResponse[QueryArgs, QueryState]{}, err
	}
	ownerOut := qrow.Owner
	st := req.State
	st.QueryName = qrow.Name
	st.QueryVersion = qrow.Version
	st.State = qrow.State
	st.CreatedAt = qrow.CreatedAt.Format(time.RFC3339)
	st.UpdatedAt = qrow.UpdatedAt.Format(time.RFC3339)
	st.OwnerOut = &ownerOut
	return infer.ReadResponse[QueryArgs, QueryState]{ID: req.ID, Inputs: st.QueryArgs, State: st}, nil
}

// Update supports owner change only.
// Update permits updating only the owning role of a query; SQL and relation
// topology changes require replacement.
func (Query) Update(ctx context.Context, req infer.UpdateRequest[QueryArgs, QueryState]) (infer.UpdateResponse[QueryState], error) {
	st := req.State
	if (st.Owner == nil && req.Inputs.Owner == nil) || (st.Owner != nil && req.Inputs.Owner != nil && *st.Owner == *req.Inputs.Owner) {
		return infer.UpdateResponse[QueryState]{}, nil
	}
	if req.DryRun {
		st.Owner = req.Inputs.Owner
		return infer.UpdateResponse[QueryState]{Output: st}, nil
	}
	if st.QueryID == "" {
		return infer.UpdateResponse[QueryState]{}, fmt.Errorf("missing queryId for update")
	}
	if req.Inputs.Owner == nil {
		return infer.UpdateResponse[QueryState]{}, fmt.Errorf("owner update requires non-nil owner")
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.UpdateResponse[QueryState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(st.Owner, ptr.Deref(cfg.Role, "")) // current owner
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.UpdateResponse[QueryState]{}, err
	}
	defer conn.Close()
	stmt := fmt.Sprintf("ALTER QUERY %s OWNER TO %s;", st.QueryID, *req.Inputs.Owner)
	if _, err := conn.ExecContext(ctx2, stmt); err != nil {
		return infer.UpdateResponse[QueryState]{}, fmt.Errorf("failed altering owner: %w", err)
	}
	qrow, err := lookupQuery(ctx2, conn, st.QueryID)
	if err != nil {
		return infer.UpdateResponse[QueryState]{}, err
	}
	newOwner := qrow.Owner
	st.Owner = req.Inputs.Owner
	st.OwnerOut = &newOwner
	st.UpdatedAt = qrow.UpdatedAt.Format(time.RFC3339)
	return infer.UpdateResponse[QueryState]{Output: st}, nil
}

// Delete terminates the query and waits for termination.
// Delete terminates an active query (if necessary) and waits for terminal state.
func (Query) Delete(ctx context.Context, req infer.DeleteRequest[QueryState]) (infer.DeleteResponse, error) {
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
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	defer conn.Close()
	if req.State.State != "terminated" && req.State.State != "terminate_requested" {
		term := fmt.Sprintf("TERMINATE QUERY %s;", req.ID)
		if _, err := conn.ExecContext(ctx2, term); err != nil {
			var sqlErr ds.ErrSQLError
			if !errors.As(err, &sqlErr) || sqlErr.SQLCode != ds.SqlStateInvalidQuery {
				return infer.DeleteResponse{}, err
			}
		}
	}
	_ = waitForQueryTerminated(ctx2, conn, req.ID, time.Minute*5)
	return infer.DeleteResponse{}, nil
}

// Internal helpers similar to object.go but query-focused
type queryRow struct {
	Name      *string
	Version   *int64
	State     string
	Owner     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// describeQuery executes DESCRIBE against the provided SQL text returning the
// classified statement kind and a parsed structural plan.
func describeQuery(ctx context.Context, conn *sql.Conn, sqlText string) (string, queryStatementPlan, error) {
	q := fmt.Sprintf("DESCRIBE %s", sqlText)
	row := conn.QueryRowContext(ctx, q)
	var kind, descJSON string
	if err := row.Scan(&kind, &descJSON); err != nil {
		return "", queryStatementPlan{}, err
	}
	var plan queryStatementPlan
	if err := json.Unmarshal([]byte(descJSON), &plan); err != nil {
		return "", queryStatementPlan{}, err
	}
	return kind, plan, nil
}

// executeQuerySQL runs the query creation DDL and returns the generated artifact descriptor.
func executeQuerySQL(ctx context.Context, conn *sql.Conn, sqlText string) (queryArtifactDDL, error) {
	row := conn.QueryRowContext(ctx, sqlText)
	var art queryArtifactDDL
	if err := row.Scan(&art.Type, &art.Name, &art.Command, &art.Summary, &art.Path); err != nil {
		return art, err
	}
	return art, nil
}

// lookupQuery returns a hydrated query row for the given ID.
func lookupQuery(ctx context.Context, conn *sql.Conn, id string) (queryRow, error) {
	q := fmt.Sprintf("select name, \"version\", current_state, \"owner\", created_at, updated_at from deltastream.sys.\"queries\" where id = '%s';", id)
	row := conn.QueryRowContext(ctx, q)
	var r queryRow
	if err := row.Scan(&r.Name, &r.Version, &r.State, &r.Owner, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return r, err
	}
	return r, nil
}

// ensureQueryLookup attempts a lookup and returns any encountered error.
func ensureQueryLookup(ctx context.Context, conn *sql.Conn, id string) error {
	_, err := lookupQuery(ctx, conn, id)
	return err
}

// waitForQueryRunning polls until the query reaches running or errored/timeout occurs.
func waitForQueryRunning(ctx context.Context, conn *sql.Conn, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		qr, err := lookupQuery(ctx, conn, id)
		if err != nil {
			var sqlErr ds.ErrSQLError
			if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidQuery {
				return fmt.Errorf("query disappeared during startup")
			}
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("query row missing")
			}
			// transient; brief sleep
		} else {
			switch qr.State {
			case "running":
				return nil
			case "errored":
				return fmt.Errorf("query errored while starting")
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for query to reach running state")
		}
		time.Sleep(time.Second * 5)
	}
}

// waitForQueryTerminated polls until the query reaches terminated or disappears.
func waitForQueryTerminated(ctx context.Context, conn *sql.Conn, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		qr, err := lookupQuery(ctx, conn, id)
		if err != nil {
			var sqlErr ds.ErrSQLError
			if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidQuery {
				return nil
			}
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
		} else if qr.State == "terminated" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for query termination")
		}
		time.Sleep(time.Second * 5)
	}
}

// WireDependencies associates dynamic outputs with input fields for Pulumi diff graph.
func (Query) WireDependencies(f infer.FieldSelector, args *QueryArgs, state *QueryState) {
	f.OutputField(&state.State).DependsOn(f.InputField(&args.SQL))
	f.OutputField(&state.QueryID).DependsOn(f.InputField(&args.SQL))
}

// provisionalQueryID derives a stable preview ID for planning phases.
func provisionalQueryID(in *QueryArgs) string {
	// Similar to relation provisional logic: derive a stable-ish preview ID from sink + hash of SQL length
	h := len(in.SQL)
	base := in.SinkRelationFqn
	if base == "" {
		base = "query"
	}
	return fmt.Sprintf("preview-%s-%d", sanitizePreviewID(base), h)
}

// sanitizePreviewID truncates and normalizes a string for provisional IDs.
func sanitizePreviewID(s string) string {
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, " ", "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
