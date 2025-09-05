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
	"os"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"
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
		// tlsCaCertFile ignored when TLS disabled
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
	if k.TlsVerifyServerHostname != nil && (k.TlsDisabled == nil || !*k.TlsDisabled) {
		params["tls.verify_server_hostname"] = boolToSql(*k.TlsVerifyServerHostname)
	}
	if strings.EqualFold(k.SaslHashFunction, "AWS_MSK_IAM") {
		if k.MskIamRoleArn != nil {
			params["kafka.msk.iam_role_arn"] = fmt.Sprintf("'%s'", *k.MskIamRoleArn)
		}
		if k.MskAwsRegion != nil {
			params["kafka.msk.aws_region"] = fmt.Sprintf("'%s'", *k.MskAwsRegion)
		}
	} else if strings.EqualFold(k.SaslHashFunction, "PLAIN") || strings.EqualFold(k.SaslHashFunction, "SHA512") || strings.EqualFold(k.SaslHashFunction, "SHA256") {
		if k.SaslUsername != nil {
			params["kafka.sasl.username"] = fmt.Sprintf("'%s'", *k.SaslUsername)
		}
		if k.SaslPassword != nil {
			params["kafka.sasl.password"] = fmt.Sprintf("'%s'", *k.SaslPassword)
		}
	}
	if k.SchemaRegistryName != nil {
		params["kafka.schema_registry_name"] = fmt.Sprintf("'%s'", *k.SchemaRegistryName)
	}
	if k.TlsCaCertFile != nil && (k.TlsDisabled == nil || !*k.TlsDisabled) {
		content, err := os.ReadFile(*k.TlsCaCertFile)
		if err != nil {
			return fmt.Errorf("failed reading tlsCaCertFile: %w", err)
		}
		esc := strings.ReplaceAll(string(content), "'", "''")
		params["tls.ca_cert"] = fmt.Sprintf("'%s'", esc)
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
	if curr.SaslHashFunction != old.SaslHashFunction {
		esc := strings.ReplaceAll(curr.SaslHashFunction, "'", "''")
		changes["kafka.sasl.hash_function"] = fmt.Sprintf("'%s'", esc)
	}
	isMSKNew := strings.EqualFold(curr.SaslHashFunction, "AWS_MSK_IAM")
	isMSKOld := strings.EqualFold(old.SaslHashFunction, "AWS_MSK_IAM")
	if isMSKNew {
		if !isMSKOld {
			if old.SaslUsername != nil {
				changes["kafka.sasl.username"] = "NULL"
			}
			if old.SaslPassword != nil {
				changes["kafka.sasl.password"] = "NULL"
			}
		}
		setIfChanged("kafka.msk.iam_role_arn", curr.MskIamRoleArn, old.MskIamRoleArn)
		setIfChanged("kafka.msk.aws_region", curr.MskAwsRegion, old.MskAwsRegion)
	} else {
		if isMSKOld {
			if old.MskIamRoleArn != nil {
				changes["kafka.msk.iam_role_arn"] = "NULL"
			}
			if old.MskAwsRegion != nil {
				changes["kafka.msk.aws_region"] = "NULL"
			}
		}
		setIfChanged("kafka.sasl.username", curr.SaslUsername, old.SaslUsername)
		setIfChanged("kafka.sasl.password", curr.SaslPassword, old.SaslPassword)
	}
	setIfChanged("kafka.schema_registry_name", curr.SchemaRegistryName, old.SchemaRegistryName)
	if (curr.TlsDisabled == nil) != (old.TlsDisabled == nil) || (curr.TlsDisabled != nil && old.TlsDisabled != nil && *curr.TlsDisabled != *old.TlsDisabled) {
		if curr.TlsDisabled == nil {
			changes["tls.disabled"] = "NULL"
		} else {
			changes["tls.disabled"] = boolToSql(*curr.TlsDisabled)
		}
	}
	if (curr.TlsVerifyServerHostname == nil) != (old.TlsVerifyServerHostname == nil) || (curr.TlsVerifyServerHostname != nil && old.TlsVerifyServerHostname != nil && *curr.TlsVerifyServerHostname != *old.TlsVerifyServerHostname) {
		if curr.TlsVerifyServerHostname == nil {
			changes["tls.verify_server_hostname"] = "NULL"
		} else if curr.TlsDisabled == nil || !*curr.TlsDisabled {
			changes["tls.verify_server_hostname"] = boolToSql(*curr.TlsVerifyServerHostname)
		}
	}
	if (curr.TlsCaCertFile == nil) != (old.TlsCaCertFile == nil) || (curr.TlsCaCertFile != nil && old.TlsCaCertFile != nil && *curr.TlsCaCertFile != *old.TlsCaCertFile) {
		if curr.TlsCaCertFile == nil {
			changes["tls.ca_cert"] = "NULL"
		} else if curr.TlsDisabled == nil || !*curr.TlsDisabled {
			content, err := os.ReadFile(*curr.TlsCaCertFile)
			if err != nil {
				return infer.UpdateResponse[StoreState]{}, fmt.Errorf("failed reading tlsCaCertFile: %w", err)
			}
			esc := strings.ReplaceAll(string(content), "'", "''")
			changes["tls.ca_cert"] = fmt.Sprintf("'%s'", esc)
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
