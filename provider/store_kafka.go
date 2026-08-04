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
	"io"
	"os"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"

	godeltastream "github.com/deltastreaminc/go-deltastream"
)

// KafkaInputs holds Kafka-specific properties for a store (split from store.go).
type KafkaInputs struct {
	Uris                    string  `pulumi:"uris"`
	SchemaRegistryName      *string `pulumi:"schemaRegistryName,optional"`
	SaslHashFunction        string  `pulumi:"saslHashFunction"`
	SaslUsername            *string `pulumi:"saslUsername,optional" provider:"secret"`
	SaslPassword            *string `pulumi:"saslPassword,optional" provider:"secret"`
	MskIamRoleArn           *string `pulumi:"mskIamRoleArn,optional"`
	MskAwsRegion            *string `pulumi:"mskAwsRegion,optional"`
	TlsDisabled             *bool   `pulumi:"tlsDisabled,optional"`
	TlsVerifyServerHostname *bool   `pulumi:"tlsVerifyServerHostname,optional"`
	TlsCaCertFile           *string `pulumi:"tlsCaCertFile,optional"`
}

// Validation for Kafka specific configuration

// validateKafkaInputs provides reusable validation logic for KafkaInputs without requiring a full CheckRequest.
// storeKafkaCheck (legacy single-type path) and multi-dispatch Store.Check both rely on this helper now.
func validateKafkaInputs(k *KafkaInputs) []p.CheckFailure {
	failures := []p.CheckFailure{}
	if k == nil {
		failures = append(failures, p.CheckFailure{Property: "kafka", Reason: "kafka block required"})
		return failures
	}
	if k.SaslHashFunction == "" {
		failures = append(failures, p.CheckFailure{Property: "kafka.saslHashFunction", Reason: "saslHashFunction required"})
	}
	if k.Uris == "" {
		failures = append(failures, p.CheckFailure{Property: "kafka.uris", Reason: "uris required"})
	}
	isMSK := strings.EqualFold(k.SaslHashFunction, "AWS_MSK_IAM")
	isSASL := strings.EqualFold(k.SaslHashFunction, "PLAIN") || strings.EqualFold(k.SaslHashFunction, "SHA512") || strings.EqualFold(k.SaslHashFunction, "SHA256")
	if isMSK {
		if k.MskIamRoleArn == nil || *k.MskIamRoleArn == "" {
			failures = append(failures, p.CheckFailure{Property: "kafka.mskIamRoleArn", Reason: "mskIamRoleArn required when saslHashFunction=AWS_MSK_IAM"})
		}
		if k.MskAwsRegion == nil || *k.MskAwsRegion == "" {
			failures = append(failures, p.CheckFailure{Property: "kafka.mskAwsRegion", Reason: "mskAwsRegion required when saslHashFunction=AWS_MSK_IAM"})
		}
		if (k.SaslUsername != nil && *k.SaslUsername != "") || (k.SaslPassword != nil && *k.SaslPassword != "") {
			failures = append(failures, p.CheckFailure{Property: "kafka.saslUsername", Reason: "saslUsername/password not allowed for AWS_MSK_IAM"})
		}
	} else if isSASL {
		if k.SaslUsername == nil || *k.SaslUsername == "" {
			failures = append(failures, p.CheckFailure{Property: "kafka.saslUsername", Reason: "saslUsername required for SCRAM mode"})
		}
		if k.SaslPassword == nil || *k.SaslPassword == "" {
			failures = append(failures, p.CheckFailure{Property: "kafka.saslPassword", Reason: "saslPassword required for SCRAM mode"})
		}
	}
	if k.TlsDisabled != nil && *k.TlsDisabled {
		// tlsCaCertFile is ignored when TLS is disabled; no action required.
		_ = k.TlsDisabled
	}
	return failures
}

// storeKafkaCreate issues CREATE STORE for a Kafka store.
func storeKafkaCreate(ctx context.Context, conn *sql.Conn, input *StoreArgs) error {
	k := input.Kafka
	params := map[string]string{
		"kafka.sasl.hash_function": k.SaslHashFunction,
		"uris":                     fmt.Sprintf("'%s'", k.Uris),
	}
	if k.TlsDisabled != nil {
		params["tls.disabled"] = boolToSql(*k.TlsDisabled)
		if *k.TlsDisabled {
			params["tls.verify_server_hostname"] = "FALSE"
		}
	}
	if k.TlsDisabled == nil || !*k.TlsDisabled {
		// Default to TRUE (the secure default) when unset.
		verifyServerHostname := k.TlsVerifyServerHostname == nil || *k.TlsVerifyServerHostname
		params["tls.verify_server_hostname"] = boolToSql(verifyServerHostname)
	}
	if strings.EqualFold(k.SaslHashFunction, "AWS_MSK_IAM") {
		setStringIfPresent(params, "kafka.msk.iam_role_arn", k.MskIamRoleArn)
		setStringIfPresent(params, "kafka.msk.aws_region", k.MskAwsRegion)
	} else if strings.EqualFold(k.SaslHashFunction, "PLAIN") || strings.EqualFold(k.SaslHashFunction, "SHA512") || strings.EqualFold(k.SaslHashFunction, "SHA256") {
		setStringIfPresent(params, "kafka.sasl.username", k.SaslUsername)
		setStringIfPresent(params, "kafka.sasl.password", k.SaslPassword)
	}
	setStringIfPresent(params, "kafka.schema_registry_name", k.SchemaRegistryName)
	if k.TlsCaCertFile != nil && (k.TlsDisabled == nil || !*k.TlsDisabled) {
		content, err := os.ReadFile(*k.TlsCaCertFile)
		if err != nil {
			return fmt.Errorf("failed reading tlsCaCertFile: %w", err)
		}
		params["tls.ca_cert_file"] = "'@cacert'"
		ctx = godeltastream.WithAttachment(ctx, "@cacert", io.NopCloser(strings.NewReader(string(content))))
	}
	pairs := make([]string, 0, len(params))
	for kkey, v := range params {
		if v == "" {
			continue
		}
		if v == "TRUE" || v == "FALSE" {
			pairs = append(pairs, fmt.Sprintf("'%s' = %s", kkey, v))
		} else {
			pairs = append(pairs, fmt.Sprintf("'%s' = %s", kkey, v))
		}
	}
	stmt := fmt.Sprintf("CREATE STORE %s WITH ( 'type' = KAFKA, %s );", quoteIdent(input.Name), strings.Join(pairs, ", "))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	return nil
}

// storeKafkaUpdate applies in-place changes to a Kafka store.
func storeKafkaUpdate(ctx context.Context, req infer.UpdateRequest[StoreArgs, StoreState]) (infer.UpdateResponse[StoreState], error) {
	if req.DryRun {
		return infer.UpdateResponse[StoreState]{}, nil
	}
	input := req.Inputs
	prev := req.State.StoreArgs
	if input.Kafka == nil || prev.Kafka == nil {
		return infer.UpdateResponse[StoreState]{}, fmt.Errorf("update currently only supports kafka inputs")
	}
	curr := input.Kafka
	old := prev.Kafka
	changes := map[string]string{}
	setIfChanged := func(key string, newPtr *string, oldPtr *string) {
		newVal := ""
		if newPtr != nil {
			newVal = *newPtr
		}
		oldVal := ""
		if oldPtr != nil {
			oldVal = *oldPtr
		}
		if newPtr == nil && oldPtr != nil {
			changes[key] = "NULL"
			return
		}
		if newPtr != nil && newVal != oldVal {
			if newVal == "" {
				changes[key] = "NULL"
			} else {
				esc := strings.ReplaceAll(newVal, "'", "''")
				changes[key] = fmt.Sprintf("'%s'", esc)
			}
		}
	}
	if curr.Uris != old.Uris {
		esc := strings.ReplaceAll(curr.Uris, "'", "''")
		changes["uris"] = fmt.Sprintf("'%s'", esc)
	}
	// The backend re-verifies the store's connectivity/auth on every UPDATE
	// STORE, and it validates the SASL/TLS parameters as a complete group
	// rather than as an independent diff. If any auth-related field is sent
	// without its siblings (e.g. changing only the CA cert while leaving the
	// SASL credentials out) the verification fails with errors such as
	// "kafka.sasl.password, and kafka.sasl.username are required".
	//
	// To avoid that, always re-send the full set of credential parameters
	// that are relevant for the current hash function, plus the TLS/CA
	// parameters, regardless of whether they individually changed. The exact
	// set of required parameters depends on kafka.sasl.hash_function (see the
	// UPDATE STORE reference docs):
	//
	//   AWS_MSK_IAM        -> kafka.msk.iam_role_arn, kafka.msk.aws_region
	//   PLAIN/SHA256/SHA512 -> kafka.sasl.username, kafka.sasl.password
	//   NONE               -> (no credentials)
	//
	// kafka.sasl.hash_function is a bareword enum token (not a quoted string
	// literal), matching storeKafkaCreate's handling.
	changes["kafka.sasl.hash_function"] = curr.SaslHashFunction

	// kafka.msk.iam_role_arn, kafka.msk.aws_region, kafka.sasl.username and
	// kafka.sasl.password are all STRING_VALUE-typed parameters in the
	// UPDATE STORE grammar: they do not accept a bare NULL literal. Instead
	// of forcing them to NULL when they don't apply to the current auth
	// mode, simply omit them from the WITH (...) clause via
	// setStringIfPresent (see conn.go).
	isMSK := strings.EqualFold(curr.SaslHashFunction, "AWS_MSK_IAM")
	isSASL := strings.EqualFold(curr.SaslHashFunction, "PLAIN") ||
		strings.EqualFold(curr.SaslHashFunction, "SHA256") ||
		strings.EqualFold(curr.SaslHashFunction, "SHA512")
	if isMSK {
		// MSK IAM auth: role ARN + region are required; SASL creds don't apply.
		setStringIfPresent(changes, "kafka.msk.iam_role_arn", curr.MskIamRoleArn)
		setStringIfPresent(changes, "kafka.msk.aws_region", curr.MskAwsRegion)
	} else if isSASL {
		// SCRAM/PLAIN auth: username + password are required; MSK params don't apply.
		setStringIfPresent(changes, "kafka.sasl.username", curr.SaslUsername)
		setStringIfPresent(changes, "kafka.sasl.password", curr.SaslPassword)
	}
	// NONE: no SASL credentials or MSK params apply; nothing to set.

	setIfChanged("kafka.schema_registry_name", curr.SchemaRegistryName, old.SchemaRegistryName)

	// Always re-send the TLS parameters so the backend re-validates the
	// connection with a complete, self-consistent TLS configuration. The
	// cert-file parameters (tls.ca_cert_file, and likewise
	// tls.client.cert_file / tls.client.key_file) are optional and are only
	// sent when the user provides them.
	// tls.disabled and tls.verify_server_hostname are boolean-typed
	// parameters in the UPDATE STORE grammar: they only accept the bareword
	// literals TRUE/FALSE, not a bare NULL. When unset, tls.disabled
	// defaults to FALSE (its implicit default) while
	// tls.verify_server_hostname defaults to TRUE (the secure default)
	// instead of sending an invalid NULL literal.
	tlsDisabled := curr.TlsDisabled != nil && *curr.TlsDisabled
	changes["tls.disabled"] = boolToSql(tlsDisabled)
	if tlsDisabled {
		// When TLS is disabled the verify/CA parameters are meaningless.
		changes["tls.verify_server_hostname"] = "FALSE"
	} else {
		verifyServerHostname := curr.TlsVerifyServerHostname == nil || *curr.TlsVerifyServerHostname
		changes["tls.verify_server_hostname"] = boolToSql(verifyServerHostname)
		// tls.ca_cert_file (like tls.client.cert_file / tls.client.key_file) is
		// optional and should only be sent when the user actually provides a
		// path. Never force it to NULL, otherwise an update that touches other
		// fields would clear a CA cert the store already relies on.
		if curr.TlsCaCertFile != nil && *curr.TlsCaCertFile != "" {
			content, err := os.ReadFile(*curr.TlsCaCertFile)
			if err != nil {
				return infer.UpdateResponse[StoreState]{}, fmt.Errorf("failed reading tlsCaCertFile: %w", err)
			}
			changes["tls.ca_cert_file"] = "'@cacert'"
			ctx = godeltastream.WithAttachment(ctx, "@cacert", io.NopCloser(strings.NewReader(string(content))))
		}
	}
	if len(changes) == 0 {
		return infer.UpdateResponse[StoreState]{}, nil
	}
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	defer db.Close() //nolint:errcheck
	role := ptr.Deref(input.Owner, ptr.Deref(cfg.Role, ""))
	org := ptr.Deref(cfg.Organization, "")
	ctx, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.UpdateResponse[StoreState]{}, err
	}
	defer conn.Close() //nolint:errcheck
	parts := make([]string, 0, len(changes))
	for k, v := range changes {
		parts = append(parts, fmt.Sprintf("'%s' = %s", k, v))
	}
	stmt := fmt.Sprintf("UPDATE STORE %s WITH ( %s );", quoteIdent(req.ID), strings.Join(parts, ", "))
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return infer.UpdateResponse[StoreState]{}, fmt.Errorf("failed updating store: %w", err)
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

// storeKafkaDiff produces a diff for Kafka-specific properties.
func storeKafkaDiff(ctx context.Context, req infer.DiffRequest[StoreArgs, StoreState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Name != req.Inputs.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.State.Kafka != nil && req.Inputs.Kafka != nil {
		s := req.State.Kafka
		n := req.Inputs.Kafka
		if s.Uris != n.Uris {
			diff["kafka.uris"] = p.PropertyDiff{Kind: p.Update}
		}
		if s.SaslHashFunction != n.SaslHashFunction {
			diff["kafka.saslHashFunction"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.SaslUsername == nil) != (n.SaslUsername == nil) || (s.SaslUsername != nil && n.SaslUsername != nil && *s.SaslUsername != *n.SaslUsername) {
			diff["kafka.saslUsername"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.SaslPassword == nil) != (n.SaslPassword == nil) || (s.SaslPassword != nil && n.SaslPassword != nil && *s.SaslPassword != *n.SaslPassword) {
			diff["kafka.saslPassword"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.MskIamRoleArn == nil) != (n.MskIamRoleArn == nil) || (s.MskIamRoleArn != nil && n.MskIamRoleArn != nil && *s.MskIamRoleArn != *n.MskIamRoleArn) {
			diff["kafka.mskIamRoleArn"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.MskAwsRegion == nil) != (n.MskAwsRegion == nil) || (s.MskAwsRegion != nil && n.MskAwsRegion != nil && *s.MskAwsRegion != *n.MskAwsRegion) {
			diff["kafka.mskAwsRegion"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.TlsDisabled == nil) != (n.TlsDisabled == nil) || (s.TlsDisabled != nil && n.TlsDisabled != nil && *s.TlsDisabled != *n.TlsDisabled) {
			diff["kafka.tlsDisabled"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.TlsVerifyServerHostname == nil) != (n.TlsVerifyServerHostname == nil) || (s.TlsVerifyServerHostname != nil && n.TlsVerifyServerHostname != nil && *s.TlsVerifyServerHostname != *n.TlsVerifyServerHostname) {
			diff["kafka.tlsVerifyServerHostname"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.TlsCaCertFile == nil) != (n.TlsCaCertFile == nil) || (s.TlsCaCertFile != nil && n.TlsCaCertFile != nil && *s.TlsCaCertFile != *n.TlsCaCertFile) {
			diff["kafka.tlsCaCertFile"] = p.PropertyDiff{Kind: p.Update}
		}
		if (s.SchemaRegistryName == nil) != (n.SchemaRegistryName == nil) || (s.SchemaRegistryName != nil && n.SchemaRegistryName != nil && *s.SchemaRegistryName != *n.SchemaRegistryName) {
			diff["kafka.schemaRegistryName"] = p.PropertyDiff{Kind: p.Update}
		}
	}
	return infer.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}
