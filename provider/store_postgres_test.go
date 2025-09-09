package provider

import (
	"context"
	"strings"
	"testing"
)

func TestNormalizePostgresUris(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		want        string
		expectError string
	}{
		{
			name:        "empty input",
			input:       "   ",
			expectError: "postgres uris cannot be empty",
		},
		{
			name:        "invalid uri parse",
			input:       "://bad",
			expectError: "invalid postgres uri",
		},
		{
			name:        "missing database name",
			input:       "example.com",
			expectError: "postgres uri must include a database name",
		},
		{
			name:  "multiple uris only first used",
			input: "host1/db1/extra,host2/db2",
			// first path segment kept (db1) and default port added
			want: "postgresql://host1:5432/db1",
		},
		{
			name:  "scheme and custom port preserved with truncation to first path segment",
			input: "postgresql://example.com:15432/db1/extra/stuff",
			want:  "postgresql://example.com:15432/db1",
		},
		{
			name:  "single path segment as database name",
			input: "example.com/db1",
			want:  "postgresql://example.com:5432/db1",
		},
		{
			name:  "adds scheme and default port",
			input: "myhost/mydb/extra",
			want:  "postgresql://myhost:5432/mydb",
		},
		{
			name:        "path with just a slash",
			input:       "example.com/",
			expectError: "postgres uri must include a database name",
		},
		{
			name:  "trailing slash in path",
			input: "example.com/mydb/",
			want:  "postgresql://example.com:5432/mydb",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizePostgresUris(context.Background(), tt.input)
			if tt.expectError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectError)
				}
				if !strings.Contains(err.Error(), tt.expectError) {
					t.Fatalf("error %q does not contain expected substring %q", err.Error(), tt.expectError)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			// Ensure only first URI used when multiple provided
			if strings.Contains(tt.input, ",") {
				parts := strings.Split(tt.input, ",")
				if len(parts) > 1 && strings.Contains(got, strings.TrimSpace(parts[1])) {
					t.Fatalf("second URI segment %q should not appear in result %q", parts[1], got)
				}
			}
		})
	}
}
