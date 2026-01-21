package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type KafkaCreds struct {
	SaslKafkaUris string `yaml:"saslKafkaUris"`
	SaslUser      string `yaml:"saslUser"`
	SaslPass      string `yaml:"saslPass"`
	IamKafkaUris  string `yaml:"iamKafkaUris"`
	MskRole       string `yaml:"mskRole"`
	MskRegion     string `yaml:"mskRegion"`
}

type SnowflakeCreds struct {
	SnowflakeUris          string `yaml:"snowflakeUris"`
	SnowflakeAccountId     string `yaml:"snowflakeAccountId"`
	SnowflakeRoleName      string `yaml:"snowflakeRoleName"`
	SnowflakeUsername      string `yaml:"snowflakeUsername"`
	SnowflakeWarehouseName string `yaml:"snowflakeWarehouseName"`
	SnowflakeCloudRegion   string `yaml:"snowflakeCloudRegion"`
	SnowflakeClientKey     string `yaml:"snowflakeClientKey"`
}

type PostgresCreds struct {
	PostgresUris        string `yaml:"postgresUris"`
	PostgresUsername    string `yaml:"postgresUsername"`
	PostgresPassword    string `yaml:"postgresPassword"`
	PostgresDatabase    string `yaml:"postgresDatabase"`
	PostgresCdcSlotName string `yaml:"postgresCdcSlotName"`
}

func getJSBaseOptions(t *testing.T) integration.ProgramTestOptions {
	t.Helper()
	base := getBaseOptions(t)
	baseJS := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			"@deltastream/pulumi-deltastream",
		},
	})

	return baseJS
}

func getBaseOptions(t *testing.T) integration.ProgramTestOptions {
	t.Helper()
	binPath, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Using binPath %s\n", binPath)
	return integration.ProgramTestOptions{
		LocalProviders: []integration.LocalDependency{
			{
				Package: "deltastream",
				Path:    binPath,
			},
		},
	}
}

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}

	return cwd
}

func getGoBaseOptions(t *testing.T) integration.ProgramTestOptions {
	t.Helper()
	goDepRoot := os.Getenv("PULUMI_GO_DEP_ROOT")
	if goDepRoot == "" {
		var err error
		goDepRoot, err = filepath.Abs("../..")
		require.NoError(t, err)
	}
	rootSdkPath, err := filepath.Abs("../sdk/go/pulumi-deltastream")
	require.NoError(t, err)

	base := getBaseOptions(t)
	baseJS := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			fmt.Sprintf("github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream=%s", rootSdkPath),
		},
		Env: []string{
			fmt.Sprintf("PULUMI_GO_DEP_ROOT=%s", goDepRoot),
		},
	})

	return baseJS
}

func getCredentials(t *testing.T) (string, string) {
	t.Helper()

	data, err := os.ReadFile("credentials.yaml")
	require.NoError(t, err)

	var creds struct {
		APIKey string `yaml:"apiKey"`
		Server string `yaml:"server"`
	}
	err = yaml.Unmarshal(data, &creds)
	require.NoError(t, err)

	if creds.APIKey == "" || creds.Server == "" {
		t.Skip("Skipping: requires DELTASTREAM_API_KEY and DELTASTREAM_SERVER in env or configure test harness to load credentials")
	}
	return creds.APIKey, creds.Server
}
