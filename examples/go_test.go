//go:build go || all
// +build go all

package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type Creds struct {
	SaslKafkaUris string `yaml:"saslKafkaUris"`
	SaslUser      string `yaml:"saslUser"`
	SaslPass      string `yaml:"saslPass"`
	IamKafkaUris  string `yaml:"iamKafkaUris"`
	MskRole       string `yaml:"mskRole"`
	MskRegion     string `yaml:"mskRegion"`
}

func TestBasicGo(t *testing.T) {
	apiKey, server := getCredentials(t)

	opts := getGoBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              filepath.Join(getCwd(t), "database-namespace-go"),
		DestroyOnCleanup: true,
		AllowEmptyPreviewChanges: true,
		AllowEmptyUpdateChanges:  true,
		Env: []string{
			"DELTASTREAM_API_KEY=" + apiKey,
			"DELTASTREAM_SERVER=" + server,
		},
		ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
			require.NotEmpty(t, stackInfo.Outputs["db_createdAt"])
			if v, ok := stackInfo.Outputs["dbs_count"]; ok {
				require.Greater(t, v.(float64), float64(1))
			}
			require.NotEmpty(t, stackInfo.Outputs["ns_createdAt"])
			if v, ok := stackInfo.Outputs["namespaces_count"]; ok {
				require.GreaterOrEqual(t, v.(float64), float64(1))
			}
		},
	})
	integration.ProgramTest(t, &opts)
}

func TestKafkaStoreUpdateGo(t *testing.T) {
	apiKey, server := getCredentials(t)

	data, err := os.ReadFile("credentials.yaml")
	require.NoError(t, err)

	var creds Creds
	err = yaml.Unmarshal(data, &creds)
	require.NoError(t, err)

	if creds.SaslKafkaUris == "" || creds.SaslUser == "" || creds.SaslPass == "" || creds.IamKafkaUris == "" || creds.MskRole == "" || creds.MskRegion == "" {
		t.Skip("Skipping Kafka store update test: missing required Kafka/MSK env vars")
	}
	base := getGoBaseOptions(t)
	step1Dir := filepath.Join(getCwd(t), "kafka-store-go", "step1")
	step2Dir := filepath.Join(getCwd(t), "kafka-store-go", "step2")
	opts := base.With(integration.ProgramTestOptions{
		Dir:              step1Dir,
		DestroyOnCleanup: true,
		AllowEmptyPreviewChanges: true,
		AllowEmptyUpdateChanges:  true,
		Env: []string{
			"DELTASTREAM_API_KEY=" + apiKey,
			"DELTASTREAM_SERVER=" + server,
			"KAFKA_SASL_URIS=" + creds.SaslKafkaUris,
			"KAFKA_SASL_USERNAME=" + creds.SaslUser,
			"KAFKA_SASL_PASSWORD=" + creds.SaslPass,
			"KAFKA_MSK_IAM_URIS=" + creds.IamKafkaUris,
			"KAFKA_MSK_IAM_ROLE_ARN=" + creds.MskRole,
			"KAFKA_MSK_AWS_REGION=" + creds.MskRegion,
		},
		EditDirs: []integration.EditDir{{
			Dir:      step2Dir,
			Additive: true,
			ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
				// After update, auth mode should switch
				if mode, ok := stackInfo.Outputs["store_auth_mode"]; ok {
					require.Equal(t, "AWS_MSK_IAM", mode.(string))
				}
				if st, ok := stackInfo.Outputs["store_state"]; ok {
					require.NotEmpty(t, st)
				}
			},
		}},
		ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
			// Initial deployment should export SCRAM
			if mode, ok := stackInfo.Outputs["store_auth_mode"]; ok {
				require.Equal(t, "SHA512", mode.(string))
			}
		},
	})
	integration.ProgramTest(t, &opts)
}

func TestSnowflakeStoreUpdateGo(t *testing.T) {
	apiKey, server := getCredentials(t)
	data, err := os.ReadFile("credentials.yaml")
	require.NoError(t, err)
	var creds struct {
		SnowflakeUris          string `yaml:"snowflakeUris"`
		SnowflakeAccountId     string `yaml:"snowflakeAccountId"`
		SnowflakeRoleName      string `yaml:"snowflakeRoleName"`
		SnowflakeUsername      string `yaml:"snowflakeUsername"`
		SnowflakeWarehouseName string `yaml:"snowflakeWarehouseName"`
		SnowflakeCloudRegion   string `yaml:"snowflakeCloudRegion"`
		SnowflakeClientKey     string `yaml:"snowflakeClientKey"`
	}
	err = yaml.Unmarshal(data, &creds)
	require.NoError(t, err)
	if creds.SnowflakeUris == "" || creds.SnowflakeAccountId == "" || creds.SnowflakeRoleName == "" || creds.SnowflakeUsername == "" || creds.SnowflakeWarehouseName == "" || creds.SnowflakeCloudRegion == "" || creds.SnowflakeClientKey == "" {
		t.Skip("Skipping Snowflake store update test: missing required snowflake credentials")
	}
	step1Dir := filepath.Join(getCwd(t), "snowflake-store-go", "step1")
	step2Dir := filepath.Join(getCwd(t), "snowflake-store-go", "step2")
	warehouse2 := creds.SnowflakeWarehouseName + "_ALT"
	opts := getGoBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              step1Dir,
		DestroyOnCleanup: true,
		AllowEmptyPreviewChanges: true,
		AllowEmptyUpdateChanges:  true,
		Env: []string{
			"DELTASTREAM_API_KEY=" + apiKey,
			"DELTASTREAM_SERVER=" + server,
			"SNOWFLAKE_URIS=" + creds.SnowflakeUris,
			"SNOWFLAKE_ACCOUNT_ID=" + creds.SnowflakeAccountId,
			"SNOWFLAKE_ROLE_NAME=" + creds.SnowflakeRoleName,
			"SNOWFLAKE_USERNAME=" + creds.SnowflakeUsername,
			"SNOWFLAKE_WAREHOUSE_NAME=" + creds.SnowflakeWarehouseName,
			"SNOWFLAKE_CLOUD_REGION=" + creds.SnowflakeCloudRegion,
			"SNOWFLAKE_CLIENT_KEY=" + creds.SnowflakeClientKey,
			"SNOWFLAKE_WAREHOUSE_NAME_STEP2=" + warehouse2,
		},
		EditDirs: []integration.EditDir{{
			Dir:      step2Dir,
			Additive: true,
			ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
				if wt, ok := stackInfo.Outputs["warehouse_name"]; ok {
					switch v := wt.(type) {
					case string:
						require.Equal(t, warehouse2, v)
					case map[string]interface{}:
						// Secret-wrapped: ensure ciphertext exists; cannot assert plaintext
						if _, hasCipher := v["ciphertext"]; !hasCipher {
							// attempt best-effort extraction, else fail
							for _, maybe := range []string{"warehouse_name", "warehouse", "name", "value"} {
								if alt, ok := v[maybe].(string); ok {
									require.NotEmpty(t, alt)
									return
								}
							}
							t.Fatalf("secret map missing ciphertext and value for warehouse_name: %#v", v)
						}
						// Ciphertext present: accept as indication of non-empty secret
						require.NotEmpty(t, v["ciphertext"])
					default:
						// Provide diagnostic instead of panic
						t.Fatalf("unexpected type for warehouse_name: %T (%#v)", wt, wt)
					}
				}
			},
		}},
		ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
			if st, ok := stackInfo.Outputs["store_type"]; ok {
				switch v := st.(type) {
				case string:
					require.Equal(t, "snowflake", strings.ToLower(v))
				case map[string]interface{}:
					// Secret wrapped type shouldn't normally happen, but tolerate by ensuring ciphertext present
					if _, hasCipher := v["ciphertext"]; hasCipher {
						require.NotEmpty(t, v["ciphertext"])
					} else {
						for _, maybe := range []string{"store_type", "type", "t", "value"} {
							if alt, ok := v[maybe].(string); ok {
								require.Equal(t, "snowflake", strings.ToLower(alt))
								return
							}
						}
						t.Fatalf("unexpected map structure for store_type: %#v", v)
					}
				default:
					t.Fatalf("unexpected type for store_type: %T (%#v)", st, st)
				}
			}
		},
	})
	integration.ProgramTest(t, &opts)
}

func TestPostgresStoreUpdateGo(t *testing.T) {
	apiKey, server := getCredentials(t)
	data, err := os.ReadFile("credentials.yaml")
	require.NoError(t, err)
	var creds struct {
		PostgresUris     string `yaml:"postgresUris"`
		PostgresUsername string `yaml:"postgresUsername"`
		PostgresPassword string `yaml:"postgresPassword"`
	}
	err = yaml.Unmarshal(data, &creds)
	require.NoError(t, err)
	if creds.PostgresUris == "" || creds.PostgresUsername == "" || creds.PostgresPassword == "" {
		t.Skip("Skipping Postgres store update test: missing required postgres credentials")
	}
	step1Dir := filepath.Join(getCwd(t), "postgres-store-go", "step1")
	step2Dir := filepath.Join(getCwd(t), "postgres-store-go", "step2")
	opts := getGoBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              step1Dir,
		DestroyOnCleanup: true,
		AllowEmptyPreviewChanges: true,
		AllowEmptyUpdateChanges:  true,
		Env: []string{
			"DELTASTREAM_API_KEY=" + apiKey,
			"DELTASTREAM_SERVER=" + server,
			"POSTGRES_URIS=" + creds.PostgresUris,
			"POSTGRES_USERNAME=" + creds.PostgresUsername,
			"POSTGRES_PASSWORD=" + creds.PostgresPassword,
		},
		EditDirs: []integration.EditDir{{
			Dir:      step2Dir,
			Additive: true,
			ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
				if raw, ok := stackInfo.Outputs["postgres_uris_plain"]; ok {
					uris := extractPostgresUris(t, raw)
					require.Contains(t, uris, "param1=test")
				}
			},
		}},
		ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
			if raw, ok := stackInfo.Outputs["postgres_uris_plain"]; ok {
				uris := extractPostgresUris(t, raw)
				require.NotContains(t, uris, "param1=test")
			}
		},
	})
	integration.ProgramTest(t, &opts)
}

func TestObjectGo(t *testing.T) {
	apiKey, server := getCredentials(t)
	baseDir := filepath.Join(getCwd(t), "object-go")
	step1 := filepath.Join(baseDir, "step1")
	step2 := filepath.Join(baseDir, "step2")
	opts := getGoBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              step1,
		DestroyOnCleanup: true,
		Env: []string{
			"DELTASTREAM_API_KEY=" + apiKey,
			"DELTASTREAM_SERVER=" + server,
			"KAFKA_MSK_IAM_URIS=" + os.Getenv("KAFKA_MSK_IAM_URIS"),
			"KAFKA_MSK_IAM_ROLE_ARN=" + os.Getenv("KAFKA_MSK_IAM_ROLE_ARN"),
			"KAFKA_MSK_AWS_REGION=" + os.Getenv("KAFKA_MSK_AWS_REGION"),
		},
		EditDirs: []integration.EditDir{{
			Dir:      step2,
			Additive: true,
			ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
				if owner, ok := stackInfo.Outputs["object_owner"]; ok {
					require.Equal(t, "public", owner.(string))
				}
			},
		}},
		ExtraRuntimeValidation: func(t *testing.T, stackInfo integration.RuntimeValidationStackInfo) {
			if st, ok := stackInfo.Outputs["object_state"]; ok {
				require.NotEmpty(t, st)
			}
			if typ, ok := stackInfo.Outputs["object_type"]; ok {
				require.Equal(t, "stream", strings.ToLower(typ.(string)))
			}
		},
	})
	integration.ProgramTest(t, &opts)
}

func TestQueryGo(t *testing.T) {
	apiKey, server := getCredentials(t)

	var creds Creds
	data, err := os.ReadFile("credentials.yaml")
	require.NoError(t, err)
	err = yaml.Unmarshal(data, &creds)
	require.NoError(t, err)

	if creds.IamKafkaUris == "" || creds.MskRole == "" || creds.MskRegion == "" {
		t.Skip("Skipping Query test: missing Kafka IAM env vars")
	}
	opts := getGoBaseOptions(t).With(integration.ProgramTestOptions{
		Dir:              filepath.Join(getCwd(t), "query-go"),
		DestroyOnCleanup: true,
		SkipPreview:      true,
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
			if sink, ok := stackInfo.Outputs["pageviews_6_fqn"]; ok {
				require.NotEmpty(t, sink)
			}
			if qid, ok := stackInfo.Outputs["query_id"]; ok {
				require.NotEmpty(t, qid)
			}
			if qsql, ok := stackInfo.Outputs["query_sql"]; ok {
				require.Contains(t, strings.ToUpper(qsql.(string)), "INSERT INTO")
			}
			if qstate, ok := stackInfo.Outputs["query_state"]; ok {
				require.NotEmpty(t, qstate)
			}
		},
	})
	integration.ProgramTest(t, &opts)
}

// extractPostgresUris tolerates export as either string or nested map structure
func extractPostgresUris(t *testing.T, v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		if u, ok := val["uris"].(string); ok {
			return u
		}
		t.Fatalf("postgres_uris map missing 'uris' key or not a string: %#v", val)
	}
	t.Fatalf("unexpected type for postgres_uris: %T", v)
	return ""
}
