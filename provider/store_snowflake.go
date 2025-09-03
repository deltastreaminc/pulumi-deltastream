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
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"

	godeltastream "github.com/deltastreaminc/go-deltastream"
)

// SnowflakeInputs holds Snowflake-specific store configuration. All fields are
// required and ClientKey must be a base64 encoded private key material which is
// attached out-of-band using an in-memory attachment labelled @keyfile.
type SnowflakeInputs struct {
	Uris          string `pulumi:"uris"`
	AccountId     string `pulumi:"accountId"`
	RoleName      string `pulumi:"roleName"`
	Username      string `pulumi:"username"`
	WarehouseName string `pulumi:"warehouseName"`
	CloudRegion   string `pulumi:"cloudRegion"`
	ClientKey     string `pulumi:"clientKey" provider:"secret"` // base64 encoded private key
}

// validateSnowflakeInputs ensures required Snowflake attributes are present and
// performs basic base64 validation of the client key.
func validateSnowflakeInputs(s *SnowflakeInputs) []p.CheckFailure {
	failures := []p.CheckFailure{}
	if s == nil {
		failures = append(failures, p.CheckFailure{Property: "snowflake", Reason: "snowflake block required"})
		return failures
	}
	if s.Uris == "" {
		failures = append(failures, p.CheckFailure{Property: "snowflake.uris", Reason: "uris required"})
	}
	if s.AccountId == "" {
		failures = append(failures, p.CheckFailure{Property: "snowflake.accountId", Reason: "accountId required"})
	}
	if s.RoleName == "" {
		failures = append(failures, p.CheckFailure{Property: "snowflake.roleName", Reason: "roleName required"})
	}
	if s.Username == "" {
		failures = append(failures, p.CheckFailure{Property: "snowflake.username", Reason: "username required"})
	}
	if s.WarehouseName == "" {
		failures = append(failures, p.CheckFailure{Property: "snowflake.warehouseName", Reason: "warehouseName required"})
	}
	if s.CloudRegion == "" {
		failures = append(failures, p.CheckFailure{Property: "snowflake.cloudRegion", Reason: "cloudRegion required"})
	}
	if s.ClientKey == "" {
		failures = append(failures, p.CheckFailure{Property: "snowflake.clientKey", Reason: "clientKey (base64) required"})
	} else {
		// basic base64 validation
		if _, err := base64.StdEncoding.DecodeString(s.ClientKey); err != nil {
			failures = append(failures, p.CheckFailure{Property: "snowflake.clientKey", Reason: "clientKey must be valid base64"})
		}
	}
	return failures
}

// storeSnowflakeCreate issues a CREATE STORE for Snowflake attaching the decoded
// private key via context attachment with an indirection reference.
func storeSnowflakeCreate(ctx context.Context, conn *sql.Conn, input *StoreArgs) error {
	s := input.Snowflake
	if s == nil {
		return fmt.Errorf("snowflake inputs missing")
	}
	// Decode base64 client key and attach in-memory; always use symbolic @keyfile reference.
	decoded, err := base64.StdEncoding.DecodeString(s.ClientKey)
	if err != nil {
		return fmt.Errorf("invalid snowflake clientKey base64: %w", err)
	}
	keyFileVal := "@keyfile"
	esc := func(v string) string { return strings.ReplaceAll(v, "'", "''") }
	pairs := []string{
		"'type' = SNOWFLAKE",
		fmt.Sprintf("'uris' = '%s'", esc(s.Uris)),
		fmt.Sprintf("'snowflake.account_id' = '%s'", esc(s.AccountId)),
		fmt.Sprintf("'snowflake.role_name' = '%s'", esc(s.RoleName)),
		fmt.Sprintf("'snowflake.username' = '%s'", esc(s.Username)),
		fmt.Sprintf("'snowflake.warehouse_name' = '%s'", esc(s.WarehouseName)),
		fmt.Sprintf("'snowflake.cloud.region' = '%s'", esc(s.CloudRegion)),
		fmt.Sprintf("'snowflake.client.key_file' = '%s'", esc(keyFileVal)),
	}
	stmt := fmt.Sprintf("CREATE STORE %s WITH ( %s );", quoteIdent(input.Name), strings.Join(pairs, ", "))
	ctx = godeltastream.WithAttachment(ctx, keyFileVal, io.NopCloser(strings.NewReader(string(decoded))))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("failed to create snowflake store: %w", err)
	}
	return nil
}

// storeSnowflakeUpdate performs mutable property updates for a Snowflake store.
// When the client key changes it is re-attached via context.
func storeSnowflakeUpdate(ctx context.Context, req infer.UpdateRequest[StoreArgs, StoreState]) (infer.UpdateResponse[StoreState], error) {
	if req.DryRun {
		return infer.UpdateResponse[StoreState]{}, nil
	}
	input := req.Inputs
	prev := req.State.StoreArgs
	if input.Snowflake == nil || prev.Snowflake == nil {
		return infer.UpdateResponse[StoreState]{}, fmt.Errorf("update currently only supports snowflake inputs when both prior and new are set")
	}
	curr := input.Snowflake
	old := prev.Snowflake
	changes := map[string]string{}
	esc := func(v string) string { return strings.ReplaceAll(v, "'", "''") }
	if curr.Uris != old.Uris {
		changes["uris"] = fmt.Sprintf("'%s'", esc(curr.Uris))
	}
	if curr.AccountId != old.AccountId {
		changes["snowflake.account_id"] = fmt.Sprintf("'%s'", esc(curr.AccountId))
	}
	if curr.RoleName != old.RoleName {
		changes["snowflake.role_name"] = fmt.Sprintf("'%s'", esc(curr.RoleName))
	}
	if curr.Username != old.Username {
		changes["snowflake.username"] = fmt.Sprintf("'%s'", esc(curr.Username))
	}
	if curr.WarehouseName != old.WarehouseName {
		changes["snowflake.warehouse_name"] = fmt.Sprintf("'%s'", esc(curr.WarehouseName))
	}
	if curr.CloudRegion != old.CloudRegion {
		changes["snowflake.cloud.region"] = fmt.Sprintf("'%s'", esc(curr.CloudRegion))
	}
	if curr.ClientKey != old.ClientKey {
		decoded, err := base64.StdEncoding.DecodeString(curr.ClientKey)
		if err != nil {
			return infer.UpdateResponse[StoreState]{}, fmt.Errorf("invalid snowflake clientKey base64: %w", err)
		}
		changes["snowflake.client.key_file"] = "'@keyfile'"
		ctx = godeltastream.WithAttachment(ctx, "keyfile", io.NopCloser(strings.NewReader(string(decoded))))
	}
	if len(changes) == 0 {
		return infer.UpdateResponse[StoreState]{}, nil
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	defer db.Close()
	role := ptr.Deref(input.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	defer conn.Close()
	parts := make([]string, 0, len(changes))
	for k, v := range changes {
		parts = append(parts, fmt.Sprintf("'%s' = %s", k, v))
	}
	stmt := fmt.Sprintf("UPDATE STORE %s WITH ( %s );", quoteIdent(req.ID), strings.Join(parts, ", "))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return infer.UpdateResponse[StoreState]{}, fmt.Errorf("failed updating snowflake store: %w", err)
	}
	sr, err := lookupStore(ctx, conn, req.ID)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	newState := req.State
	newState.StoreArgs = input
	newState.Type = sr.Type
	newState.State = sr.State
	newState.CreatedAt = sr.CreatedAt.Format(time.RFC3339)
	newState.UpdatedAt = sr.UpdatedAt.Format(time.RFC3339)
	newState.OwnerOut = sr.Owner
	return infer.UpdateResponse[StoreState]{Output: newState}, nil
}

// storeSnowflakeDiff produces a Pulumi detailed diff of Snowflake properties.
func storeSnowflakeDiff(ctx context.Context, req infer.DiffRequest[StoreArgs, StoreState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Name != req.Inputs.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.State.Snowflake != nil && req.Inputs.Snowflake != nil {
		s := req.State.Snowflake
		n := req.Inputs.Snowflake
		if s.Uris != n.Uris {
			diff["snowflake.uris"] = p.PropertyDiff{Kind: p.Update}
		}
		if s.AccountId != n.AccountId {
			diff["snowflake.accountId"] = p.PropertyDiff{Kind: p.Update}
		}
		if s.RoleName != n.RoleName {
			diff["snowflake.roleName"] = p.PropertyDiff{Kind: p.Update}
		}
		if s.Username != n.Username {
			diff["snowflake.username"] = p.PropertyDiff{Kind: p.Update}
		}
		if s.WarehouseName != n.WarehouseName {
			diff["snowflake.warehouseName"] = p.PropertyDiff{Kind: p.Update}
		}
		if s.CloudRegion != n.CloudRegion {
			diff["snowflake.cloudRegion"] = p.PropertyDiff{Kind: p.Update}
		}
		if s.ClientKey != n.ClientKey {
			diff["snowflake.clientKey"] = p.PropertyDiff{Kind: p.Update}
		}
	}
	return infer.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}
