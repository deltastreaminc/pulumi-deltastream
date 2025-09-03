//go:build nodejs || all
// +build nodejs all

package examples

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
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
