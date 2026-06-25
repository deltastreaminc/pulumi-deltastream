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
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// TestApplicationCheck_SkipsDescribeWhenSQLUnchanged verifies that Application.Check returns no
// failures and does not attempt to open a DB connection when the SQL is identical to the prior
// deploy (OldInputs["sql"] == new SQL). This prevents spurious failures caused by stale
// RESUME FROM QUERY ID references that have been garbage-collected after 30 days.
func TestApplicationCheck_SkipsDescribeWhenSQLUnchanged(t *testing.T) {
	t.Parallel()

	const sql = `BEGIN APPLICATION myapp
  INSERT INTO "db"."schema"."sink" SELECT * FROM "db"."schema"."src";
END APPLICATION WITH (resume.from.query.id = '0fe1f533-0d9d-4250-8841-8ffc47af4d71');`

	tests := []struct {
		name            string
		oldSQL          string // empty means no OldInputs["sql"] key
		newSQL          string
		wantSkip        bool // true → no failures expected (DESCRIBE skipped)
		sources         []string
		sinks           []string
	}{
		{
			name:     "unchanged SQL skips DESCRIBE",
			oldSQL:   sql,
			newSQL:   sql,
			wantSkip: true,
			sources:  []string{`"db"."schema"."src"`},
			sinks:    []string{`"db"."schema"."sink"`},
		},
		{
			name:     "no OldInputs (first create) does not skip",
			oldSQL:   "",
			newSQL:   sql,
			wantSkip: false,
			sources:  []string{`"db"."schema"."src"`},
			sinks:    []string{`"db"."schema"."sink"`},
		},
		{
			name:     "changed SQL does not skip",
			oldSQL:   sql + " -- modified",
			newSQL:   sql,
			wantSkip: false,
			sources:  []string{`"db"."schema"."src"`},
			sinks:    []string{`"db"."schema"."sink"`},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			oldInputs := property.Map{}
			if tt.oldSQL != "" {
				oldInputs = property.NewMap(map[string]property.Value{
					"sql": property.New(tt.oldSQL),
				})
			}

			newInputs := property.NewMap(map[string]property.Value{
				"sql":                property.New(tt.newSQL),
				"sourceRelationFqns": property.New(property.NewArray(func() []property.Value {
					vals := make([]property.Value, len(tt.sources))
					for i, s := range tt.sources {
						vals[i] = property.New(s)
					}
					return vals
				}())),
				"sinkRelationFqns": property.New(property.NewArray(func() []property.Value {
					vals := make([]property.Value, len(tt.sinks))
					for i, s := range tt.sinks {
						vals[i] = property.New(s)
					}
					return vals
				}())),
			})

			if tt.wantSkip {
				// When SQL is unchanged the provider must return no failures without
				// reaching the DB (no DB credentials are set so any DB attempt would fail).
				resp, err := Application{}.Check(context.Background(), infer.CheckRequest{
					OldInputs: oldInputs,
					NewInputs: newInputs,
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(resp.Failures) != 0 {
					t.Errorf("expected no failures when SQL unchanged, got: %v", resp.Failures)
				}
			} else {
				// When SQL changes or OldInputs is absent the provider proceeds past the guard
				// and attempts to open a DB connection (which panics or errors without a
				// provider context). A panic here confirms that DESCRIBE was not skipped.
				panicked := func() (p bool) {
					defer func() {
						if r := recover(); r != nil {
							p = true
						}
					}()
					_, _ = Application{}.Check(context.Background(), infer.CheckRequest{
						OldInputs: oldInputs,
						NewInputs: newInputs,
					})
					return false
				}()
				if !panicked {
					t.Error("expected DB access (panic) when SQL is changed or new, but Check returned without error")
				}
			}
		})
	}
}
