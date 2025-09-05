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
	"reflect"
	"testing"
)

// TestSecretFieldMetadata ensures that struct tags declare provider:"secret" for expected fields.
func TestSecretFieldMetadata(t *testing.T) {
	// KafkaInputs
	kt := reflect.TypeOf(KafkaInputs{})
	checkTag(t, kt, "SaslUsername")
	checkTag(t, kt, "SaslPassword")

	// SnowflakeInputs
	st := reflect.TypeOf(SnowflakeInputs{})
	checkTag(t, st, "ClientKey")

	// PostgresInputs
	pt := reflect.TypeOf(PostgresInputs{})
	checkTag(t, pt, "Password")
}

func checkTag(t *testing.T, typ reflect.Type, field string) {
	f, ok := typ.FieldByName(field)
	if !ok {
		t.Fatalf("field %s not found on %s", field, typ.Name())
	}
	tag := f.Tag.Get("provider")
	if tag != "secret" {
		t.Fatalf("expected %s.%s to have provider:\"secret\" tag, got %q", typ.Name(), field, tag)
	}
}
