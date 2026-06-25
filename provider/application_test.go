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

// TestApplicationCheck_SkipsDescribeWhenInputsUnchanged verifies that Application.Check returns
// no failures and does not attempt a DB connection when all inputs (SQL, sinks, sources) match
// the prior deploy. This prevents spurious failures from stale RESUME FROM QUERY ID references
// that have been garbage-collected after 30 days while the application is still running.
func TestApplicationCheck_SkipsDescribeWhenInputsUnchanged(t *testing.T) {
	t.Parallel()

	const sql = `BEGIN APPLICATION myapp
  INSERT INTO "db"."schema"."sink" SELECT * FROM "db"."schema"."src";
END APPLICATION WITH (resume.from.query.id = '0fe1f533-0d9d-4250-8841-8ffc47af4d71');`

	const src = `"db"."schema"."src"`
	const sink = `"db"."schema"."sink"`

	buildInputs := func(sqlStr string, sources, sinks []string) property.Map {
		srcVals := make([]property.Value, len(sources))
		for i, s := range sources {
			srcVals[i] = property.New(s)
		}
		sinkVals := make([]property.Value, len(sinks))
		for i, s := range sinks {
			sinkVals[i] = property.New(s)
		}
		return property.NewMap(map[string]property.Value{
			"sql":                property.New(sqlStr),
			"sourceRelationFqns": property.New(property.NewArray(srcVals)),
			"sinkRelationFqns":   property.New(property.NewArray(sinkVals)),
		})
	}

	newInputs := buildInputs(sql, []string{src}, []string{sink})

	tests := []struct {
		name      string
		oldInputs property.Map
		wantSkip  bool
	}{
		{
			name:      "all inputs unchanged skips DESCRIBE",
			oldInputs: buildInputs(sql, []string{src}, []string{sink}),
			wantSkip:  true,
		},
		{
			name:      "no OldInputs (first create) does not skip",
			oldInputs: property.Map{},
			wantSkip:  false,
		},
		{
			name:      "changed SQL does not skip",
			oldInputs: buildInputs(sql+" -- modified", []string{src}, []string{sink}),
			wantSkip:  false,
		},
		{
			name:      "SQL same but sinks changed does not skip",
			oldInputs: buildInputs(sql, []string{src}, []string{`"db"."schema"."old_sink"`}),
			wantSkip:  false,
		},
		{
			name:      "SQL same but sources changed does not skip",
			oldInputs: buildInputs(sql, []string{`"db"."schema"."old_src"`}, []string{sink}),
			wantSkip:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantSkip {
				// When all inputs are unchanged the provider must return no failures without
				// reaching the DB (no credentials configured so any DB attempt would fail).
				resp, err := Application{}.Check(context.Background(), infer.CheckRequest{
					OldInputs: tt.oldInputs,
					NewInputs: newInputs,
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(resp.Failures) != 0 {
					t.Errorf("expected no failures when inputs unchanged, got: %v", resp.Failures)
				}
			} else {
				// When inputs change or OldInputs is absent the provider proceeds past the
				// guard and attempts to open a DB connection, which panics without a provider
				// context. A panic here confirms DESCRIBE was not skipped.
				panicked := func() (p bool) {
					defer func() {
						if r := recover(); r != nil {
							p = true
						}
					}()
					_, _ = Application{}.Check(context.Background(), infer.CheckRequest{
						OldInputs: tt.oldInputs,
						NewInputs: newInputs,
					})
					return false
				}()
				if !panicked {
					t.Error("expected DB access (panic) when inputs changed or new, but Check returned without error")
				}
			}
		})
	}
}
