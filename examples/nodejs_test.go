//go:build nodejs || all
// +build nodejs all

package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBasicTs(t *testing.T) {
	apiKey, server := getCredentials(t)

	// Provide a stable session ID so that the provider's auto-generated random
	// sessionId does not cause diffs between previews/updates.
	stableSessionID := "test-session-" + time.Now().Format("20060102")

	opts := getJSBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              filepath.Join(getCwd(t), "database-namespace-ts"),
		DestroyOnCleanup: true,
		Env: []string{
			"DELTASTREAM_API_KEY=" + apiKey,
			"DELTASTREAM_SERVER=" + server,
			"DELTASTREAM_SESSION_ID=" + stableSessionID,
		},
	})

	integration.ProgramTest(t, &opts)
}

func TestQueryTs(t *testing.T) {
	apiKey, server := getCredentials(t)

	var creds KafkaCreds
	data, err := os.ReadFile("credentials.yaml")
	require.NoError(t, err)
	err = yaml.Unmarshal(data, &creds)
	require.NoError(t, err)

	if creds.IamKafkaUris == "" || creds.MskRole == "" || creds.MskRegion == "" {
		t.Skip("Skipping Query test: missing Kafka IAM env vars")
	}

	step1Dir := filepath.Join(getCwd(t), "query-ts", "step1")
	step2Dir := filepath.Join(getCwd(t), "query-ts", "step2")

	var step1QueryID string
	var step1PageviewsCreatedAt string

	opts := getJSBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:                      step1Dir,
		DestroyOnCleanup:         true,
		SkipPreview:              true,
		AllowEmptyPreviewChanges: true,
		AllowEmptyUpdateChanges:  true,
		Env: []string{
			"DELTASTREAM_API_KEY=" + apiKey,
			"DELTASTREAM_SERVER=" + server,
			"KAFKA_MSK_IAM_URIS=" + creds.IamKafkaUris,
			"KAFKA_MSK_IAM_ROLE_ARN=" + creds.MskRole,
			"KAFKA_MSK_AWS_REGION=" + creds.MskRegion,
		},
		ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
			if pv, ok := stackInfo.Outputs["pageviews_fqn"]; ok {
				require.NotEmpty(t, pv)
			}
			if pvCreated, ok := stackInfo.Outputs["pageviews_created_at"]; ok {
				require.NotEmpty(t, pvCreated)
				step1PageviewsCreatedAt = pvCreated.(string)
			}
			if sink, ok := stackInfo.Outputs["pageviews_6_fqn"]; ok {
				require.NotEmpty(t, sink)
			}
			if qid, ok := stackInfo.Outputs["query_id"]; ok {
				require.NotEmpty(t, qid)
				step1QueryID = qid.(string)
			}
			if qsql, ok := stackInfo.Outputs["query_sql"]; ok {
				require.Contains(t, strings.ToUpper(qsql.(string)), "INSERT INTO")
			}
			if qstate, ok := stackInfo.Outputs["query_state"]; ok {
				require.NotEmpty(t, qstate)
			}
		},
		EditDirs: []integration.EditDir{{
			Dir:      step2Dir,
			Additive: true,
			ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
				// After update, the stream SQL should have changed (WITH clause reordered)
				if pv, ok := stackInfo.Outputs["pageviews_fqn"]; ok {
					require.NotEmpty(t, pv)
				}
				if pvCreated, ok := stackInfo.Outputs["pageviews_created_at"]; ok {
					require.NotEmpty(t, pvCreated)
					step2PageviewsCreatedAt := pvCreated.(string)
					// Stream should be recreated with a new timestamp
					require.NotEqual(t, step1PageviewsCreatedAt, step2PageviewsCreatedAt, "Pageviews createdAt should change after stream recreation")
				}
				if qid, ok := stackInfo.Outputs["query_id"]; ok {
					require.NotEmpty(t, qid)
					step2QueryID := qid.(string)
					// Query should be recreated with a new ID
					require.NotEqual(t, step1QueryID, step2QueryID, "Query ID should change after stream recreation")
				}
				if qsql, ok := stackInfo.Outputs["query_sql"]; ok {
					require.Contains(t, strings.ToUpper(qsql.(string)), "INSERT INTO")
				}
				if qstate, ok := stackInfo.Outputs["query_state"]; ok {
					require.NotEmpty(t, qstate)
				}
			},
		}},
	})
	integration.ProgramTest(t, &opts)
}
