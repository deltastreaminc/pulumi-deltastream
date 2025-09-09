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
	"fmt"
	"net/url"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"
)

// PostgresInputs holds configuration for a PostgreSQL store. The URIs field may contain
// a comma-separated list of hosts (with or without explicit scheme/port). Ports are
// normalized to 5432 where absent.
type PostgresInputs struct {
	Uris                    string `pulumi:"uris"`
	Username                string `pulumi:"username"`
	Password                string `pulumi:"password" provider:"secret"`
	TlsDisabled             *bool  `pulumi:"tlsDisabled,optional"`
	TlsVerifyServerHostname *bool  `pulumi:"tlsVerifyServerHostname,optional"`
}

// validatePostgresInputs performs input validation returning Pulumi check failures.
func validatePostgresInputs(pg *PostgresInputs) []p.CheckFailure {
	var failures []p.CheckFailure
	if pg == nil {
		return failures
	}
	if pg.Uris == "" {
		failures = append(failures, p.CheckFailure{Property: "postgres.uris", Reason: "uris is required"})
	}
	if pg.Username == "" {
		failures = append(failures, p.CheckFailure{Property: "postgres.username", Reason: "username is required"})
	}
	if pg.Password == "" {
		failures = append(failures, p.CheckFailure{Property: "postgres.password", Reason: "password is required"})
	}
	return failures
}

// storePostgresCreate issues a CREATE STORE statement for Postgres including
// normalization of host URIs and TLS flags.
func storePostgresCreate(ctx context.Context, conn *sql.Conn, input *StoreArgs) (err error) {
	pg := input.Postgres
	// normalize URIs (comma separated) to ensure each host has a port
	pg.Uris, err = normalizePostgresUris(ctx, pg.Uris)
	if err != nil {
		return err
	}
	params := []string{"'type' = POSTGRESQL"}
	params = append(params, fmt.Sprintf("'postgres.username' = '%s'", escapeSQL(pg.Username)))
	params = append(params, fmt.Sprintf("'postgres.password' = '%s'", escapeSQL(pg.Password)))
	params = append(params, fmt.Sprintf("'uris' = '%s'", escapeSQL(pg.Uris)))
	if pg.TlsDisabled != nil {
		if *pg.TlsDisabled {
			params = append(params, "'tls.disabled' = TRUE")
			// if disabled we force verify_server_hostname false
			params = append(params, "'tls.verify_server_hostname' = FALSE")
		} else {
			params = append(params, "'tls.disabled' = FALSE")
		}
	}
	if pg.TlsVerifyServerHostname != nil && (pg.TlsDisabled == nil || !*pg.TlsDisabled) {
		if *pg.TlsVerifyServerHostname {
			params = append(params, "'tls.verify_server_hostname' = TRUE")
		} else {
			params = append(params, "'tls.verify_server_hostname' = FALSE")
		}
	}
	stmt := fmt.Sprintf("CREATE STORE %s WITH ( %s );", quoteIdent(input.Name), strings.Join(params, ", "))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

// storePostgresUpdate performs in-place updates of mutable Postgres properties
// (username, password, uris, tls flags) and waits for store readiness after mutation.
func storePostgresUpdate(ctx context.Context, req infer.UpdateRequest[StoreArgs, StoreState]) (infer.UpdateResponse[StoreState], error) {
	// Open connection similar to other update helpers
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(req.State.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	defer conn.Close()

	changes := map[string]string{}
	if req.Inputs.Postgres.Username != req.State.Postgres.Username {
		changes["postgres.username"] = quoteString(req.Inputs.Postgres.Username)
	}
	if req.Inputs.Postgres.Password != req.State.Postgres.Password {
		changes["postgres.password"] = quoteString(req.Inputs.Postgres.Password)
	}
	// normalize both sides for fair comparison
	newUris, err := normalizePostgresUris(ctx, req.Inputs.Postgres.Uris)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	oldUris, err := normalizePostgresUris(ctx, req.State.Postgres.Uris)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	if newUris != oldUris {
		changes["uris"] = quoteString(newUris)
		req.Inputs.Postgres.Uris = newUris
	} else {
		req.Inputs.Postgres.Uris = oldUris
	}
	// TLS booleans
	curTD, oldTD := req.Inputs.Postgres.TlsDisabled, req.State.Postgres.TlsDisabled
	if (curTD == nil) != (oldTD == nil) || (curTD != nil && oldTD != nil && *curTD != *oldTD) {
		if curTD == nil {
			changes["tls.disabled"] = "FALSE"
		} else if *curTD {
			changes["tls.disabled"] = "TRUE"
			// force verify_server_hostname false when disabled
			changes["tls.verify_server_hostname"] = "FALSE"
		} else {
			changes["tls.disabled"] = "FALSE"
		}
	}
	curV, oldV := req.Inputs.Postgres.TlsVerifyServerHostname, req.State.Postgres.TlsVerifyServerHostname
	if (curV == nil) != (oldV == nil) || (curV != nil && oldV != nil && *curV != *oldV) {
		if curV == nil {
			changes["tls.verify_server_hostname"] = "NULL"
		} else if req.Inputs.Postgres.TlsDisabled == nil || (req.Inputs.Postgres.TlsDisabled != nil && !*req.Inputs.Postgres.TlsDisabled) {
			if *curV {
				changes["tls.verify_server_hostname"] = "TRUE"
			} else {
				changes["tls.verify_server_hostname"] = "FALSE"
			}
		}
	}
	if len(changes) > 0 {
		parts := []string{}
		for k, v := range changes {
			parts = append(parts, fmt.Sprintf("'%s' = %s", k, escapeSQL(v)))
		}
		stmt := fmt.Sprintf("UPDATE STORE %s WITH (%s);", quoteIdent(req.ID), joinComma(parts))
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			return infer.UpdateResponse[StoreState]{}, err
		}
		// wait ready
		if _, err := waitForStoreReady(ctx, conn, req.ID); err != nil {
			return infer.UpdateResponse[StoreState]{}, err
		}
	}
	st := req.State
	st.Postgres = req.Inputs.Postgres
	return infer.UpdateResponse[StoreState]{Output: st}, nil
}

// storePostgresDiff calculates detailed property differences for Postgres-specific
// settings to drive plan output without replacement.
func storePostgresDiff(ctx context.Context, req infer.DiffRequest[StoreArgs, StoreState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.Postgres.Username != req.State.Postgres.Username {
		diff["postgres.username"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.Postgres.Password != req.State.Postgres.Password {
		diff["postgres.password"] = p.PropertyDiff{Kind: p.Update}
	}
	old, err := normalizePostgresUris(ctx, req.State.Postgres.Uris)
	if err != nil {
		return infer.DiffResponse{}, err
	}
	new, err := normalizePostgresUris(ctx, req.Inputs.Postgres.Uris)
	if err != nil {
		return infer.DiffResponse{}, err
	}
	if new != old {
		diff["postgres.uris"] = p.PropertyDiff{Kind: p.Update}
	}
	if (req.Inputs.Postgres.TlsDisabled == nil) != (req.State.Postgres.TlsDisabled == nil) || (req.Inputs.Postgres.TlsDisabled != nil && req.State.Postgres.TlsDisabled != nil && *req.Inputs.Postgres.TlsDisabled != *req.State.Postgres.TlsDisabled) {
		diff["postgres.tlsDisabled"] = p.PropertyDiff{Kind: p.Update}
	}
	if (req.Inputs.Postgres.TlsVerifyServerHostname == nil) != (req.State.Postgres.TlsVerifyServerHostname == nil) || (req.Inputs.Postgres.TlsVerifyServerHostname != nil && req.State.Postgres.TlsVerifyServerHostname != nil && *req.Inputs.Postgres.TlsVerifyServerHostname != *req.State.Postgres.TlsVerifyServerHostname) {
		diff["postgres.tlsVerifyServerHostname"] = p.PropertyDiff{Kind: p.Update}
	}
	if len(diff) == 0 {
		return infer.DiffResponse{HasChanges: false}, nil
	}
	return infer.DiffResponse{HasChanges: true, DetailedDiff: diff}, nil
}

// joinComma joins parts with commas without creating an intermediate slice via strings.Join.
// helper functions reused
func joinComma(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += ", " + parts[i]
	}
	return out
}

// escapeSQL performs the most basic single-quote escaping used for embedding values
// inside single-quoted SQL literal contexts.
func escapeSQL(s string) string { return strings.ReplaceAll(s, "'", "''") }

func normalizePostgresUris(ctx context.Context, in string) (string, error) {
	logger := p.GetLogger(ctx)
	if strings.TrimSpace(in) == "" {
		return "", fmt.Errorf("postgres uris cannot be empty")
	}
	parts := strings.Split(in, ",")
	if len(parts) > 1 {
		logger.Info("postgres store was provided multiple uris; using the first one only")
	}

	seg := strings.TrimSpace(parts[0])
	if !strings.Contains(seg, "://") { // ensure scheme
		seg = "postgresql://" + seg
	}

	u, err := url.Parse(seg)
	if err != nil || u.Host == "" { // leave as-is on parse failure
		return "", fmt.Errorf("invalid postgres uri: %w", err)
	}
	host := u.Host
	if u.Port() == "" {
		host = u.Hostname() + ":5432"
	}
	u.Host = host

	// Path should contain the database name (e.g., /mydb)
	if u.Path == "" || u.Path == "/" {
		return "", fmt.Errorf("postgres uri must include a database name")
	}

	// Remove leading and trailing slashes and split by slash to get path segments
	trimmedPath := strings.Trim(u.Path, "/")
	if trimmedPath == "" {
		return "", fmt.Errorf("postgres uri must include a database name")
	}

	segments := strings.Split(trimmedPath, "/")

	// First segment is the database name
	dbName := segments[0]
	if len(segments) > 1 {
		logger.Info(fmt.Sprintf("postgres uri contained multiple path segments; using %s as database name", dbName))
	}

	// Set the path to just the database name
	u.Path = "/" + dbName

	return u.String(), nil
}
