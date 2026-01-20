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

// Application resource implements multi-sink APPLICATION queries.
type Application struct{}

func (a *Application) Annotate(an infer.Annotator) {
	an.Describe(a, "Application resource for multi-sink streaming applications with virtual intermediate relations.")
}

// ApplicationArgs define the desired APPLICATION query expressed as a
// BEGIN APPLICATION ... END APPLICATION statement along with explicit
// enumeration of source and sink relations.
type ApplicationArgs struct {
	SourceRelationFqns []string `pulumi:"sourceRelationFqns"`
	SinkRelationFqns   []string `pulumi:"sinkRelationFqns"`
	SQL                string   `pulumi:"sql"`
	Owner              *string  `pulumi:"owner,optional"`
}

// ApplicationState captures runtime attributes of an APPLICATION after creation.
type ApplicationState struct {
	ApplicationArgs
	ApplicationID string  `pulumi:"applicationId"`
	QueryName     *string `pulumi:"queryName,optional"`
	QueryVersion  *int64  `pulumi:"queryVersion,optional"`
	State         string  `pulumi:"state"`
	CreatedAt     string  `pulumi:"createdAt"`
	UpdatedAt     string  `pulumi:"updatedAt"`
	OwnerOut      *string `pulumi:"owner"`
}

func (s *ApplicationState) Annotate(a infer.Annotator) {
	a.Describe(&s.State, "Lifecycle state of the application (starting|running|terminate_requested|terminated|errored)")
	a.Describe(&s.ApplicationID, "System-generated application identifier")
}

// applicationRelationPlan represents a relation with is_virtual flag
type applicationRelationPlan struct {
	Fqn        string `json:"fqn"`
	Type       string `json:"type"`
	DbName     string `json:"db_name"`
	SchemaName string `json:"schema_name"`
	Name       string `json:"name"`
	StoreName  string `json:"store_name"`
	IsVirtual  bool   `json:"is_virtual"`
}

// applicationStatementPlan structures the DESCRIBE APPLICATION output
type applicationStatementPlan struct {
	Ddls       []applicationRelationPlan `json:"ddls,omitempty"`
	Statements []string                  `json:"statements,omitempty"`
	Sources    []applicationRelationPlan `json:"sources,omitempty"`
	Sinks      []applicationRelationPlan `json:"sinks,omitempty"`
}

// isVirtualRelation checks if a relation is virtual (not backed by a physical store)
func isVirtualRelation(rel applicationRelationPlan) bool {
	return rel.IsVirtual
}

// Check validates the APPLICATION SQL by running DESCRIBE and verifying plan alignment.
func (Application) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[ApplicationArgs], error) {
	args, failures, err := infer.DefaultCheck[ApplicationArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[ApplicationArgs]{}, err
	}

	if args.SQL == "" || len(args.SinkRelationFqns) == 0 || len(args.SourceRelationFqns) == 0 {
		return infer.CheckResponse[ApplicationArgs]{Inputs: args, Failures: failures}, nil
	}

	cfg := infer.GetConfig[Config](ctx)
	db, oerr := openDB(ctx, &cfg)
	if oerr != nil { // tolerate missing connection in preview
		return infer.CheckResponse[ApplicationArgs]{Inputs: args, Failures: failures}, nil
	}
	defer db.Close()

	role := ptr.Deref(args.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, cerr := withOrgRole(ctx, db, org, role)
	if cerr != nil {
		return infer.CheckResponse[ApplicationArgs]{Inputs: args, Failures: failures}, nil
	}
	defer conn.Close()

	kind, plan, derr := describeApplication(ctx2, conn, args.SQL)
	if derr != nil {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("describe failed: %v", derr)})
		return infer.CheckResponse[ApplicationArgs]{Inputs: args, Failures: failures}, nil
	}

	if kind != "APPLICATION" {
		failures = append(failures, p.CheckFailure{Property: "sql", Reason: fmt.Sprintf("unsupported query kind: %s (only APPLICATION queries are supported)", kind)})
		return infer.CheckResponse[ApplicationArgs]{Inputs: args, Failures: failures}, nil
	}

	// Validate ALL physical sinks are declared
	for _, sink := range plan.Sinks {
		sinkFqn := getFQN([]string{sink.DbName, sink.SchemaName, sink.Name})
		found := false
		for _, declared := range args.SinkRelationFqns {
			if strings.TrimSpace(declared) == sinkFqn {
				found = true
				break
			}
		}
		if !found {
			failures = append(failures, p.CheckFailure{Property: "sinkRelationFqns", Reason: fmt.Sprintf("missing sink %s", sink.Fqn)})
		}
	}

	// Validate no extra sinks are declared
	for _, declared := range args.SinkRelationFqns {
		found := false
		for _, sink := range plan.Sinks {
			if strings.TrimSpace(declared) == getFQN([]string{sink.DbName, sink.SchemaName, sink.Name}) {
				found = true
				break
			}
		}
		if !found {
			sinkFqns := make([]string, len(plan.Sinks))
			for i, s := range plan.Sinks {
				sinkFqns[i] = s.Fqn
			}
			failures = append(failures, p.CheckFailure{Property: "sinkRelationFqns", Reason: fmt.Sprintf("declared sink %s not found in application sinks: %v", declared, sinkFqns)})
		}
	}

	// Validate that all declared sources are present in the physical sources
	for _, declared := range args.SourceRelationFqns {
		found := false
		for _, src := range plan.Sources {
			if strings.TrimSpace(declared) == getFQN([]string{src.DbName, src.SchemaName, src.Name}) {
				found = true
				break
			}
		}
		if !found {
			failures = append(failures, p.CheckFailure{Property: "sourceRelationFqns", Reason: fmt.Sprintf("declared source %s not found in application physical sources", declared)})
		}
	}

	// Ensure every physical source in the application is declared
	for _, src := range plan.Sources {
		found := false
		for _, declared := range args.SourceRelationFqns {
			if strings.TrimSpace(declared) == getFQN([]string{src.DbName, src.SchemaName, src.Name}) {
				found = true
				break
			}
		}
		if !found {
			failures = append(failures, p.CheckFailure{Property: "sourceRelationFqns", Reason: fmt.Sprintf("missing physical source %s", src.Fqn)})
		}
	}

	// Check for virtual relations in sourceRelationFqns (should be rejected)
	for _, ddl := range plan.Ddls {
		if isVirtualRelation(ddl) {
			ddlFqn := getFQN([]string{ddl.DbName, ddl.SchemaName, ddl.Name})
			for _, declared := range args.SourceRelationFqns {
				if strings.TrimSpace(declared) == ddlFqn {
					failures = append(failures, p.CheckFailure{Property: "sourceRelationFqns", Reason: fmt.Sprintf("virtual relation %s cannot be declared as a source dependency", ddl.Fqn)})
				}
			}
		}
	}

	return infer.CheckResponse[ApplicationArgs]{Inputs: args, Failures: failures}, nil
}

// Diff: all fields replace except owner (update)
func (Application) Diff(ctx context.Context, req infer.DiffRequest[ApplicationArgs, ApplicationState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	if req.State.SQL != "" && req.State.SQL != req.Inputs.SQL {
		diff["sql"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	if len(req.State.SinkRelationFqns) > 0 && !stringSlicesEqual(req.State.SinkRelationFqns, req.Inputs.SinkRelationFqns) {
		diff["sinkRelationFqns"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	if len(req.State.SourceRelationFqns) > 0 && !stringSlicesEqual(req.State.SourceRelationFqns, req.Inputs.SourceRelationFqns) {
		diff["sourceRelationFqns"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	if (req.State.Owner == nil && req.Inputs.Owner != nil) || (req.State.Owner != nil && req.Inputs.Owner != nil && *req.State.Owner != *req.Inputs.Owner) {
		diff["owner"] = p.PropertyDiff{Kind: p.Update}
	}

	return infer.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}

// Create launches the application and waits until running
func (Application) Create(ctx context.Context, req infer.CreateRequest[ApplicationArgs]) (infer.CreateResponse[ApplicationState], error) {
	in := req.Inputs
	logger := p.GetLogger(ctx)

	if req.DryRun {
		now := time.Now().UTC().Format(time.RFC3339)
		st := ApplicationState{ApplicationArgs: in, CreatedAt: now, UpdatedAt: now, State: "starting"}
		return infer.CreateResponse[ApplicationState]{ID: provisionalApplicationID(&in), Output: st}, nil
	}

	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.CreateResponse[ApplicationState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(in.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.CreateResponse[ApplicationState]{}, err
	}
	defer conn.Close()

	kind, plan, derr := describeApplication(ctx2, conn, in.SQL)
	if derr != nil {
		return infer.CreateResponse[ApplicationState]{}, derr
	}

	if kind != "APPLICATION" {
		return infer.CreateResponse[ApplicationState]{}, fmt.Errorf("unsupported query kind: %s (only APPLICATION queries are supported)", kind)
	}

	// Validate ALL physical sinks are declared
	for _, sink := range plan.Sinks {
		sinkFqn := getFQN([]string{sink.DbName, sink.SchemaName, sink.Name})
		found := false
		for _, declared := range in.SinkRelationFqns {
			if strings.TrimSpace(declared) == sinkFqn {
				found = true
				break
			}
		}
		if !found {
			return infer.CreateResponse[ApplicationState]{}, fmt.Errorf("physical sink relation %s not declared", sink.Fqn)
		}
	}

	// Validate no extra sinks are declared
	for _, declared := range in.SinkRelationFqns {
		found := false
		for _, sink := range plan.Sinks {
			if strings.TrimSpace(declared) == getFQN([]string{sink.DbName, sink.SchemaName, sink.Name}) {
				found = true
				break
			}
		}
		if !found {
			physicalSinkFqns := make([]string, len(plan.Sinks))
			for i, s := range plan.Sinks {
				physicalSinkFqns[i] = s.Fqn
			}
			return infer.CreateResponse[ApplicationState]{}, fmt.Errorf("declared sink %s not found in application sinks: %v", declared, physicalSinkFqns)
		}
	}

	// Validate that all declared sources are present in the physical sources
	for _, declared := range in.SourceRelationFqns {
		found := false
		for _, src := range plan.Sources {
			if strings.TrimSpace(declared) == getFQN([]string{src.DbName, src.SchemaName, src.Name}) {
				found = true
				break
			}
		}
		if !found {
			return infer.CreateResponse[ApplicationState]{}, fmt.Errorf("declared source %s not found in application physical sources", declared)
		}
	}

	// Ensure every physical source in the application is declared
	for _, src := range plan.Sources {
		found := false
		for _, declared := range in.SourceRelationFqns {
			if strings.TrimSpace(declared) == getFQN([]string{src.DbName, src.SchemaName, src.Name}) {
				found = true
				break
			}
		}
		if !found {
			return infer.CreateResponse[ApplicationState]{}, fmt.Errorf("physical source relation %s not declared", src.Fqn)
		}
	}

	// Check for virtual relations in sourceRelationFqns (should be rejected)
	for _, ddl := range plan.Ddls {
		if isVirtualRelation(ddl) {
			ddlFqn := getFQN([]string{ddl.DbName, ddl.SchemaName, ddl.Name})
			for _, declared := range in.SourceRelationFqns {
				if strings.TrimSpace(declared) == ddlFqn {
					return infer.CreateResponse[ApplicationState]{}, fmt.Errorf("virtual relation %s cannot be declared as a source dependency", ddl.Fqn)
				}
			}
		}
	}

	// Execute the APPLICATION SQL
	art, aerr := executeQuerySQL(ctx2, conn, in.SQL)
	if aerr != nil {
		return infer.CreateResponse[ApplicationState]{}, aerr
	}

	appID := art.Name

	// Poll for running state
	if perr := waitForQueryRunning(ctx2, conn, appID, time.Minute*10); perr != nil {
		_ = ensureQueryLookup(ctx2, conn, appID) // best-effort
		return infer.CreateResponse[ApplicationState]{}, perr
	}

	qrow, err := lookupQuery(ctx2, conn, appID)
	if err != nil {
		return infer.CreateResponse[ApplicationState]{}, err
	}

	ownerOut := qrow.Owner
	st := ApplicationState{
		ApplicationArgs: in,
		ApplicationID:   appID,
		QueryName:       qrow.Name,
		QueryVersion:    qrow.Version,
		State:           qrow.State,
		CreatedAt:       qrow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       qrow.UpdatedAt.Format(time.RFC3339),
		OwnerOut:        &ownerOut,
	}

	logger.Info(fmt.Sprintf("Application created: %s", appID))
	return infer.CreateResponse[ApplicationState]{ID: appID, Output: st}, nil
}

// Read refreshes state
func (Application) Read(ctx context.Context, req infer.ReadRequest[ApplicationArgs, ApplicationState]) (infer.ReadResponse[ApplicationArgs, ApplicationState], error) {
	if req.ID == "" {
		return infer.ReadResponse[ApplicationArgs, ApplicationState]{}, nil
	}

	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.ReadResponse[ApplicationArgs, ApplicationState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.ReadResponse[ApplicationArgs, ApplicationState]{}, err
	}
	defer conn.Close()

	qrow, err := lookupQuery(ctx2, conn, req.ID)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidQuery {
			return infer.ReadResponse[ApplicationArgs, ApplicationState]{}, nil
		}
		if errors.Is(err, sql.ErrNoRows) {
			return infer.ReadResponse[ApplicationArgs, ApplicationState]{}, nil
		}
		return infer.ReadResponse[ApplicationArgs, ApplicationState]{}, err
	}

	ownerOut := qrow.Owner
	st := req.State
	st.QueryName = qrow.Name
	st.QueryVersion = qrow.Version
	st.State = qrow.State
	st.CreatedAt = qrow.CreatedAt.Format(time.RFC3339)
	st.UpdatedAt = qrow.UpdatedAt.Format(time.RFC3339)
	st.OwnerOut = &ownerOut

	return infer.ReadResponse[ApplicationArgs, ApplicationState]{ID: req.ID, Inputs: st.ApplicationArgs, State: st}, nil
}

// Update supports owner change only
func (Application) Update(ctx context.Context, req infer.UpdateRequest[ApplicationArgs, ApplicationState]) (infer.UpdateResponse[ApplicationState], error) {
	st := req.State

	if (st.Owner == nil && req.Inputs.Owner == nil) || (st.Owner != nil && req.Inputs.Owner != nil && *st.Owner == *req.Inputs.Owner) {
		return infer.UpdateResponse[ApplicationState]{}, nil
	}

	if req.DryRun {
		st.Owner = req.Inputs.Owner
		return infer.UpdateResponse[ApplicationState]{Output: st}, nil
	}

	if st.ApplicationID == "" {
		return infer.UpdateResponse[ApplicationState]{}, fmt.Errorf("missing applicationId for update")
	}

	if req.Inputs.Owner == nil {
		return infer.UpdateResponse[ApplicationState]{}, fmt.Errorf("owner update requires non-nil owner")
	}

	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.UpdateResponse[ApplicationState]{}, err
	}
	defer db.Close()

	role := ptr.Deref(st.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.UpdateResponse[ApplicationState]{}, err
	}
	defer conn.Close()

	stmt := fmt.Sprintf("ALTER QUERY %s OWNER TO %s;", st.ApplicationID, *req.Inputs.Owner)
	if _, err := conn.ExecContext(ctx2, stmt); err != nil {
		return infer.UpdateResponse[ApplicationState]{}, fmt.Errorf("failed altering owner: %w", err)
	}

	qrow, err := lookupQuery(ctx2, conn, st.ApplicationID)
	if err != nil {
		return infer.UpdateResponse[ApplicationState]{}, err
	}

	newOwner := qrow.Owner
	st.Owner = req.Inputs.Owner
	st.OwnerOut = &newOwner
	st.UpdatedAt = qrow.UpdatedAt.Format(time.RFC3339)

	return infer.UpdateResponse[ApplicationState]{Output: st}, nil
}

// Delete terminates the application and waits for termination
func (Application) Delete(ctx context.Context, req infer.DeleteRequest[ApplicationState]) (infer.DeleteResponse, error) {
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

// Internal helpers for application management

// describeApplication executes DESCRIBE against the provided APPLICATION SQL text
func describeApplication(ctx context.Context, conn *sql.Conn, sqlText string) (string, applicationStatementPlan, error) {
	q := fmt.Sprintf("DESCRIBE %s", sqlText)
	row := conn.QueryRowContext(ctx, q)

	var kind, descJSON string
	if err := row.Scan(&kind, &descJSON); err != nil {
		return "", applicationStatementPlan{}, err
	}

	var plan applicationStatementPlan
	if err := json.Unmarshal([]byte(descJSON), &plan); err != nil {
		return "", applicationStatementPlan{}, err
	}

	return kind, plan, nil
}

// WireDependencies associates dynamic outputs with input fields for Pulumi diff graph
func (Application) WireDependencies(f infer.FieldSelector, args *ApplicationArgs, state *ApplicationState) {
	f.OutputField(&state.State).DependsOn(f.InputField(&args.SQL))
	f.OutputField(&state.ApplicationID).DependsOn(f.InputField(&args.SQL))
}

// provisionalApplicationID derives a stable preview ID for planning phases
func provisionalApplicationID(in *ApplicationArgs) string {
	h := len(in.SQL)
	var base string
	if len(in.SinkRelationFqns) > 0 {
		base = in.SinkRelationFqns[0]
	} else {
		base = "application"
	}
	return fmt.Sprintf("preview-%s-%d", sanitizePreviewID(base), h)
}
