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
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"

	ds "github.com/deltastreaminc/go-deltastream"
)

// readiness polling constants (spec D-012)
const (
	storeReadinessTimeout = 10 * time.Minute
	storePollInterval     = 5 * time.Second
)

// Store resource (Kafka, Postgres, Snowflake) representing an external data store.
type Store struct{}

func (s *Store) Annotate(a infer.Annotator) {
	a.Describe(s, "Store resource supporting external data store connectivity (initial Kafka support)")
}

// StoreArgs defines user inputs; exactly one subtype (Kafka, Snowflake, Postgres) must be set.
type StoreArgs struct {
	Name      string           `pulumi:"name"`
	Owner     *string          `pulumi:"owner,optional"`
	Kafka     *KafkaInputs     `pulumi:"kafka,optional"`
	Snowflake *SnowflakeInputs `pulumi:"snowflake,optional"`
	Postgres  *PostgresInputs  `pulumi:"postgres,optional"`
}

// KafkaInputs moved to store_kafka.go

// StoreState extends inputs with computed fields.
type StoreState struct {
	StoreArgs
	Type      string `pulumi:"type"`
	State     string `pulumi:"state"`
	CreatedAt string `pulumi:"createdAt"`
	UpdatedAt string `pulumi:"updatedAt"`
	OwnerOut  string `pulumi:"owner"`
}

func (s *StoreState) Annotate(a infer.Annotator) {
	a.Describe(&s.Type, "Type of the store")
	a.Describe(&s.State, "Provisioning state of the store")
}

// Check validates inputs ensuring exactly one subtype and subtype-specific validation.
func (Store) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[StoreArgs], error) {
	args, failures, err := infer.DefaultCheck[StoreArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[StoreArgs]{}, err
	}
	count := 0
	if args.Kafka != nil {
		count++
	}
	if args.Snowflake != nil {
		count++
	}
	if args.Postgres != nil {
		count++
	}
	if count == 0 {
		failures = append(failures, p.CheckFailure{Property: "", Reason: "one store subtype required (kafka, snowflake or postgres)"})
	} else if count > 1 {
		failures = append(failures, p.CheckFailure{Property: "", Reason: "only one store subtype may be specified"})
	} else {
		if args.Kafka != nil {
			failures = append(failures, validateKafkaInputs(args.Kafka)...)
		}
		if args.Snowflake != nil {
			failures = append(failures, validateSnowflakeInputs(args.Snowflake)...)
		}
		if args.Postgres != nil {
			failures = append(failures, validatePostgresInputs(args.Postgres)...)
		}
	}
	return infer.CheckResponse[StoreArgs]{Inputs: args, Failures: failures}, nil
}

// Create provisions the external store and waits until ready.
func (Store) Create(ctx context.Context, req infer.CreateRequest[StoreArgs]) (infer.CreateResponse[StoreState], error) {
	input := req.Inputs
	logger := p.GetLogger(ctx)
	logger.Debug(fmt.Sprintf("Creating store %s", input.Name))
	if req.DryRun {
		st := StoreState{StoreArgs: input, Type: "", CreatedAt: time.Now().UTC().Format(time.RFC3339)}
		if input.Kafka != nil {
			st.Type = "KAFKA"
		} else if input.Snowflake != nil {
			st.Type = "SNOWFLAKE"
		}
		return infer.CreateResponse[StoreState]{ID: input.Name, Output: st}, nil
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.CreateResponse[StoreState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(input.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.CreateResponse[StoreState]{}, err
	}
	defer conn.Close()
	switch {
	case input.Kafka != nil:
		if err := storeKafkaCreate(ctx, conn, &input); err != nil {
			return infer.CreateResponse[StoreState]{}, err
		}
	case input.Snowflake != nil:
		if err := storeSnowflakeCreate(ctx, conn, &input); err != nil {
			return infer.CreateResponse[StoreState]{}, err
		}
	case input.Postgres != nil:
		if err := storePostgresCreate(ctx, conn, &input); err != nil {
			return infer.CreateResponse[StoreState]{}, err
		}
	default:
		return infer.CreateResponse[StoreState]{}, fmt.Errorf("no store subtype provided")
	}
	sr, err := waitForStoreReady(ctx, conn, input.Name)
	if err != nil {
		return infer.CreateResponse[StoreState]{}, err
	}
	ownerOut := sr.Owner
	st := StoreState{StoreArgs: input, Type: sr.Type, State: sr.State, CreatedAt: sr.CreatedAt.Format(time.RFC3339), UpdatedAt: sr.UpdatedAt.Format(time.RFC3339), OwnerOut: ownerOut}
	logger.Info(fmt.Sprintf("Store created: %s", input.Name))
	return infer.CreateResponse[StoreState]{ID: input.Name, Output: st}, nil
}

// Read refreshes the store state from the system catalogs.
func (Store) Read(ctx context.Context, req infer.ReadRequest[StoreArgs, StoreState]) (infer.ReadResponse[StoreArgs, StoreState], error) {
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.ReadResponse[StoreArgs, StoreState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.ReadResponse[StoreArgs, StoreState]{}, err
	}
	defer conn.Close()
	sr, err := lookupStore(ctx, conn, req.ID)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidStore {
			return infer.ReadResponse[StoreArgs, StoreState]{}, nil
		}
		return infer.ReadResponse[StoreArgs, StoreState]{}, err
	}
	st := req.State
	st.Type = sr.Type
	st.State = sr.State
	st.CreatedAt = sr.CreatedAt.Format(time.RFC3339)
	st.UpdatedAt = sr.UpdatedAt.Format(time.RFC3339)
	st.OwnerOut = sr.Owner
	return infer.ReadResponse[StoreArgs, StoreState]{ID: req.ID, Inputs: st.StoreArgs, State: st}, nil
}

// Update delegates to subtype update logic when supported.
func (Store) Update(ctx context.Context, req infer.UpdateRequest[StoreArgs, StoreState]) (infer.UpdateResponse[StoreState], error) {
	if req.Inputs.Kafka != nil && req.State.Kafka != nil && req.Inputs.Snowflake == nil && req.State.Snowflake == nil {
		return storeKafkaUpdate(ctx, req)
	}
	if req.Inputs.Snowflake != nil && req.State.Snowflake != nil && req.Inputs.Kafka == nil && req.State.Kafka == nil {
		return storeSnowflakeUpdate(ctx, req)
	}
	if req.Inputs.Postgres != nil && req.State.Postgres != nil && req.Inputs.Kafka == nil && req.State.Kafka == nil && req.Inputs.Snowflake == nil && req.State.Snowflake == nil {
		return storePostgresUpdate(ctx, req)
	}
	return infer.UpdateResponse[StoreState]{}, fmt.Errorf("changing store subtype not supported; requires replacement")
}

// Delete drops the store and best-effort verifies removal.
func (Store) Delete(ctx context.Context, req infer.DeleteRequest[StoreState]) (infer.DeleteResponse, error) {
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
	stmt := fmt.Sprintf("DROP STORE %s;", quoteIdent(req.ID))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return infer.DeleteResponse{}, err
	}
	// Verify disappearance (best-effort with small timeout)
	ctxCheck, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		_, err := lookupStore(ctxCheck, conn, req.ID)
		if err != nil {
			var sqlErr ds.ErrSQLError
			if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidStore {
				// gone
				return infer.DeleteResponse{}, nil
			}
			logger := p.GetLogger(ctx)
			logger.Debug(fmt.Sprintf("unexpected error while verifying deletion of store %s: %v", req.ID, err))
		}
		select {
		case <-t.C:
		case <-ctxCheck.Done():
			// timeout - return anyway
			return infer.DeleteResponse{}, nil
		}
	}

}

// storeRow is a helper struct for querying store metadata.
type storeRow struct {
	Type, State, Owner   string
	CreatedAt, UpdatedAt time.Time
}

// lookupStore fetches basic store metadata from the catalog.
func lookupStore(ctx context.Context, conn *sql.Conn, name string) (storeRow, error) {
	q := fmt.Sprintf("SELECT type, status, \"owner\", created_at, updated_at FROM deltastream.sys.\"stores\" WHERE name = '%s';", strings.ReplaceAll(name, "'", "''"))
	row := conn.QueryRowContext(ctx, q)
	var r storeRow
	if err := row.Scan(&r.Type, &r.State, &r.Owner, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return r, err
	}
	return r, nil
}

// waitForStoreReady polls until status=ready or errored.
func waitForStoreReady(ctx context.Context, conn *sql.Conn, name string) (storeRow, error) {
	logger := p.GetLogger(ctx)
	deadline := time.Now().Add(storeReadinessTimeout)
	var last storeRow
	t := time.NewTicker(storePollInterval)
	defer t.Stop()
	for {
		if time.Now().After(deadline) {
			// attempt to fetch last status_message for richer timeout diagnostics
			msg := ""
			mrow := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT status_message FROM deltastream.sys.\"stores\" WHERE name = '%s';", strings.ReplaceAll(name, "'", "''")))
			_ = mrow.Scan(&msg)
			return last, fmt.Errorf("timeout waiting for store %s to become ready (last state=%s, lastStatusMessage=%s)", name, last.State, msg)
		}
		row := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT type, status, \"owner\", created_at, updated_at FROM deltastream.sys.\"stores\" WHERE name = '%s';", strings.ReplaceAll(name, "'", "''")))
		if err := row.Scan(&last.Type, &last.State, &last.Owner, &last.CreatedAt, &last.UpdatedAt); err != nil {
			return last, err
		}
		stateLower := strings.ToLower(last.State)
		switch stateLower {
		case "ready":
			return last, nil
		case "errored":
			msg := ""
			mrow := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT status_message FROM deltastream.sys.\"stores\" WHERE name = '%s';", strings.ReplaceAll(name, "'", "''")))
			_ = mrow.Scan(&msg)
			return last, fmt.Errorf("store %s errored: %s", name, msg)
		}
		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-t.C:
			logger.Debug(fmt.Sprintf("waiting for store %s to be ready (current=%s)", name, last.State))
		}
	}
}

func boolToSql(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

// Diff determines replacement vs in-place update across supported subtypes.
func (Store) Diff(ctx context.Context, req infer.DiffRequest[StoreArgs, StoreState]) (infer.DiffResponse, error) {
	if (req.State.Kafka != nil) != (req.Inputs.Kafka != nil) || (req.State.Snowflake != nil) != (req.Inputs.Snowflake != nil) || (req.State.Postgres != nil) != (req.Inputs.Postgres != nil) {
		diff := map[string]p.PropertyDiff{}
		diff["kafka"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff["snowflake"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff["postgres"] = p.PropertyDiff{Kind: p.UpdateReplace}
		return infer.DiffResponse{HasChanges: true, DetailedDiff: diff}, nil
	}
	if req.Inputs.Kafka != nil && req.State.Kafka != nil {
		return storeKafkaDiff(ctx, req)
	}
	if req.Inputs.Snowflake != nil && req.State.Snowflake != nil {
		return storeSnowflakeDiff(ctx, req)
	}
	if req.Inputs.Postgres != nil && req.State.Postgres != nil {
		return storePostgresDiff(ctx, req)
	}
	return infer.DiffResponse{HasChanges: false}, nil
}

// WireDependencies declares resource graph dependencies for Pulumi.
func (Store) WireDependencies(f infer.FieldSelector, inputs *StoreArgs, state *StoreState) {
	f.OutputField(&state.Type).DependsOn(f.InputField(&inputs.Kafka))
	f.OutputField(&state.Type).DependsOn(f.InputField(&inputs.Snowflake))
	f.OutputField(&state.Type).DependsOn(f.InputField(&inputs.Postgres))
	f.OutputField(&state.State).DependsOn(f.InputField(&inputs.Kafka))
	f.OutputField(&state.State).DependsOn(f.InputField(&inputs.Snowflake))
	f.OutputField(&state.State).DependsOn(f.InputField(&inputs.Postgres))
}
