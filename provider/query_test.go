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

// TestQueryCheck_SkipsDescribeWhenInputsUnchanged verifies that Query.Check returns no failures
// and does not attempt a DB connection when all inputs (SQL, sink, sources) match the prior
// deploy. This mirrors the same guard in Application.Check and guards against future additions
// of resume-style clauses in INSERT INTO queries.
func TestQueryCheck_SkipsDescribeWhenInputsUnchanged(t *testing.T) {
	t.Parallel()

	const sql = `INSERT INTO "db"."schema"."sink" SELECT * FROM "db"."schema"."src";`
	const src = `"db"."schema"."src"`
	const src2 = `"db"."schema"."src2"`
	const sink = `"db"."schema"."sink"`

	buildInputs := func(sqlStr, sinkFqn string, sources []string) property.Map {
		srcVals := make([]property.Value, len(sources))
		for i, s := range sources {
			srcVals[i] = property.New(s)
		}
		return property.NewMap(map[string]property.Value{
			"sql":                property.New(sqlStr),
			"sinkRelationFqn":    property.New(sinkFqn),
			"sourceRelationFqns": property.New(property.NewArray(srcVals)),
		})
	}

	tests := []struct {
		name      string
		oldInputs property.Map
		newInputs property.Map
		wantSkip  bool
	}{
		{
			name:      "all inputs unchanged skips DESCRIBE",
			oldInputs: buildInputs(sql, sink, []string{src}),
			newInputs: buildInputs(sql, sink, []string{src}),
			wantSkip:  true,
		},
		{
			name:      "sources in different order still skips (order-insensitive)",
			oldInputs: buildInputs(sql, sink, []string{src2, src}),
			newInputs: buildInputs(sql, sink, []string{src, src2}),
			wantSkip:  true,
		},
		{
			name:      "no OldInputs (first create) does not skip",
			oldInputs: property.Map{},
			newInputs: buildInputs(sql, sink, []string{src}),
			wantSkip:  false,
		},
		{
			name:      "changed SQL does not skip",
			oldInputs: buildInputs(sql+" -- modified", sink, []string{src}),
			newInputs: buildInputs(sql, sink, []string{src}),
			wantSkip:  false,
		},
		{
			name:      "SQL same but sink changed does not skip",
			oldInputs: buildInputs(sql, `"db"."schema"."old_sink"`, []string{src}),
			newInputs: buildInputs(sql, sink, []string{src}),
			wantSkip:  false,
		},
		{
			name:      "SQL same but sources changed does not skip",
			oldInputs: buildInputs(sql, sink, []string{`"db"."schema"."old_src"`}),
			newInputs: buildInputs(sql, sink, []string{src}),
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
				resp, err := Query{}.Check(context.Background(), infer.CheckRequest{
					OldInputs: tt.oldInputs,
					NewInputs: tt.newInputs,
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
					_, _ = Query{}.Check(context.Background(), infer.CheckRequest{
						OldInputs: tt.oldInputs,
						NewInputs: tt.newInputs,
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
