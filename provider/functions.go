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
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/pulumi/pulumi-go-provider/infer"
	"k8s.io/utils/ptr"

	ds "github.com/deltastreaminc/go-deltastream"
)

// internal cap for plural invoke enumeration (see spec D-INV-001)
const invokeMaxRows = 100

// GetDatabaseArgs defines the arguments for the GetDatabase function.
type GetDatabaseArgs struct {
	// Name of the database to look up.
	Name string `pulumi:"name"`
}

// GetDatabaseResult is the result returned from GetDatabase.
type GetDatabaseResult struct {
	// Name of the database.
	Name string `pulumi:"name"`
	// Owning role of the database.
	Owner string `pulumi:"owner"`
	// Creation timestamp in RFC3339 format.
	CreatedAt string `pulumi:"createdAt"`
}

// GetDatabase looks up a single database by name.
type GetDatabase struct{}

func (GetDatabase) Invoke(ctx context.Context, req infer.FunctionRequest[GetDatabaseArgs]) (infer.FunctionResponse[GetDatabaseResult], error) {
	args := req.Input
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetDatabaseResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetDatabaseResult]{}, err
	}
	defer conn.Close()
	owner, createdAt, err := lookupDatabase(ctx2, conn, args.Name)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidDatabase {
			return infer.FunctionResponse[GetDatabaseResult]{}, fmt.Errorf("database %s not found", args.Name)
		}
		return infer.FunctionResponse[GetDatabaseResult]{}, err
	}
	return infer.FunctionResponse[GetDatabaseResult]{Output: GetDatabaseResult{Name: args.Name, Owner: owner, CreatedAt: createdAt.Format(time.RFC3339)}}, nil
}

// GetDatabasesArgs defines the arguments for listing databases (no arguments currently).
type GetDatabasesArgs struct{}

// GetDatabasesResult contains the list of databases.
type GetDatabasesResult struct {
	// Databases returned from the system catalog.
	Databases []GetDatabaseResult `pulumi:"databases"`
}

// GetDatabases lists all databases visible to the caller.
type GetDatabases struct{}

func (GetDatabases) Invoke(ctx context.Context, req infer.FunctionRequest[GetDatabasesArgs]) (infer.FunctionResponse[GetDatabasesResult], error) {
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetDatabasesResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetDatabasesResult]{}, err
	}
	defer conn.Close()
	q := fmt.Sprintf(`SELECT name, "owner", created_at FROM deltastream.sys."databases" ORDER BY name LIMIT %d;`, invokeMaxRows+1)
	rows, err := conn.QueryContext(ctx2, q)
	if err != nil {
		return infer.FunctionResponse[GetDatabasesResult]{}, err
	}
	defer rows.Close()
	var out []GetDatabaseResult
	count := 0
	for rows.Next() {
		var name, owner string
		var created time.Time
		if err := rows.Scan(&name, &owner, &created); err != nil {
			return infer.FunctionResponse[GetDatabasesResult]{}, err
		}
		count++
		if count > invokeMaxRows { // extra sentinel row indicates truncation
			break
		}
		out = append(out, GetDatabaseResult{Name: name, Owner: owner, CreatedAt: created.Format(time.RFC3339)})
	}
	if err := rows.Err(); err != nil {
		return infer.FunctionResponse[GetDatabasesResult]{}, err
	}
	if count > invokeMaxRows {
		log.Printf("pulumi-deltastream warning: getDatabases truncated at %d rows (more available)", invokeMaxRows)
	}
	return infer.FunctionResponse[GetDatabasesResult]{Output: GetDatabasesResult{Databases: out}}, nil
}

// Namespace related invoke functions
// GetNamespaceArgs specifies the database and namespace to look up.
type GetNamespaceArgs struct {
	// Database containing the namespace.
	Database string `pulumi:"database"`
	// Namespace name.
	Name string `pulumi:"name"`
}

// GetNamespaceResult holds details about a namespace.
type GetNamespaceResult struct {
	// Database containing the namespace.
	Database string `pulumi:"database"`
	// Namespace name.
	Name string `pulumi:"name"`
	// Owning role.
	Owner string `pulumi:"owner"`
	// Creation timestamp (RFC3339).
	CreatedAt string `pulumi:"createdAt"`
}

// GetNamespace looks up a single namespace.
type GetNamespace struct{}

func (GetNamespace) Invoke(ctx context.Context, req infer.FunctionRequest[GetNamespaceArgs]) (infer.FunctionResponse[GetNamespaceResult], error) {
	args := req.Input
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetNamespaceResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetNamespaceResult]{}, err
	}
	defer conn.Close()
	owner, createdAt, err := lookupNamespace(ctx2, conn, args.Database, args.Name)
	if err != nil {
		var sqlErr ds.ErrSQLError
		if errors.As(err, &sqlErr) && sqlErr.SQLCode == ds.SqlStateInvalidSchema {
			return infer.FunctionResponse[GetNamespaceResult]{}, fmt.Errorf("namespace %s.%s not found", args.Database, args.Name)
		}
		return infer.FunctionResponse[GetNamespaceResult]{}, err
	}
	return infer.FunctionResponse[GetNamespaceResult]{Output: GetNamespaceResult{Database: args.Database, Name: args.Name, Owner: owner, CreatedAt: createdAt.Format(time.RFC3339)}}, nil
}

// GetNamespacesArgs specifies the database whose namespaces to list.
type GetNamespacesArgs struct {
	// Database name used to filter namespaces.
	Database string `pulumi:"database"`
}

// GetNamespacesResult contains namespaces within a database.
type GetNamespacesResult struct {
	// Namespaces in the database.
	Namespaces []GetNamespaceResult `pulumi:"namespaces"`
}

// GetNamespaces lists namespaces for a database.
type GetNamespaces struct{}

func (GetNamespaces) Invoke(ctx context.Context, req infer.FunctionRequest[GetNamespacesArgs]) (infer.FunctionResponse[GetNamespacesResult], error) {
	args := req.Input
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetNamespacesResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetNamespacesResult]{}, err
	}
	defer conn.Close()
	q := fmt.Sprintf(`SELECT name, "owner", created_at FROM deltastream.sys."schemas" WHERE database_name = %s ORDER BY name LIMIT %d;`, quoteString(args.Database), invokeMaxRows+1)
	rows, err := conn.QueryContext(ctx2, q)
	if err != nil {
		return infer.FunctionResponse[GetNamespacesResult]{}, err
	}
	defer rows.Close()
	var list []GetNamespaceResult
	count := 0
	for rows.Next() {
		var name, owner string
		var created time.Time
		if err := rows.Scan(&name, &owner, &created); err != nil {
			return infer.FunctionResponse[GetNamespacesResult]{}, err
		}
		count++
		if count > invokeMaxRows {
			break
		}
		list = append(list, GetNamespaceResult{Database: args.Database, Name: name, Owner: owner, CreatedAt: created.Format(time.RFC3339)})
	}
	if err := rows.Err(); err != nil {
		return infer.FunctionResponse[GetNamespacesResult]{}, err
	}
	if count > invokeMaxRows {
		log.Printf("pulumi-deltastream warning: getNamespaces truncated at %d rows (more available)", invokeMaxRows)
	}
	return infer.FunctionResponse[GetNamespacesResult]{Output: GetNamespacesResult{Namespaces: list}}, nil
}

// Store related invoke functions (initial Kafka support; generic list)
// GetStoreArgs defines the name of the store to retrieve.
type GetStoreArgs struct {
	// Store name.
	Name string `pulumi:"name"`
}

// GetStoreResult details a single store.
type GetStoreResult struct {
	// Store name.
	Name string `pulumi:"name"`
	// Store type (kafka, postgres, etc.).
	Type string `pulumi:"type"`
	// Provisioning state.
	State string `pulumi:"state"`
	// Owning role.
	Owner string `pulumi:"owner"`
	// Creation timestamp.
	CreatedAt string `pulumi:"createdAt"`
	// Last update timestamp.
	UpdatedAt string `pulumi:"updatedAt"`
}

// GetStore retrieves a single store by name.
type GetStore struct{}

func (GetStore) Invoke(ctx context.Context, req infer.FunctionRequest[GetStoreArgs]) (infer.FunctionResponse[GetStoreResult], error) {
	args := req.Input
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetStoreResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetStoreResult]{}, err
	}
	defer conn.Close()
	q := fmt.Sprintf(`SELECT type, status, "owner", created_at, updated_at FROM deltastream.sys."stores" WHERE name = %s;`, quoteString(args.Name))
	row := conn.QueryRowContext(ctx2, q)
	var typ, state, owner string
	var created, updated time.Time
	if err := row.Scan(&typ, &state, &owner, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return infer.FunctionResponse[GetStoreResult]{}, fmt.Errorf("store %s not found", args.Name)
		}
		return infer.FunctionResponse[GetStoreResult]{}, err
	}
	return infer.FunctionResponse[GetStoreResult]{Output: GetStoreResult{Name: args.Name, Type: typ, State: state, Owner: owner, CreatedAt: created.Format(time.RFC3339), UpdatedAt: updated.Format(time.RFC3339)}}, nil
}

// GetStoresArgs currently has no fields (placeholder for future filtering).
type GetStoresArgs struct{}

// GetStoresResult contains the list of stores.
type GetStoresResult struct {
	// Stores returned.
	Stores []GetStoreResult `pulumi:"stores"`
}

// GetStores lists all stores visible to the caller.
type GetStores struct{}

func (GetStores) Invoke(ctx context.Context, req infer.FunctionRequest[GetStoresArgs]) (infer.FunctionResponse[GetStoresResult], error) {
	cfg := infer.GetConfig[Config](ctx)
	db, err := openDB(ctx, &cfg)
	if err != nil {
		return infer.FunctionResponse[GetStoresResult]{}, err
	}
	defer db.Close()
	role := ptr.Deref(cfg.Role, "")
	org := ptr.Deref(cfg.Organization, "")
	ctx2, conn, err := withOrgRole(ctx, db, org, role)
	if err != nil {
		return infer.FunctionResponse[GetStoresResult]{}, err
	}
	defer conn.Close()
	q := fmt.Sprintf(`SELECT name, type, status, "owner", created_at, updated_at FROM deltastream.sys."stores" ORDER BY name LIMIT %d;`, invokeMaxRows+1)
	rows, err := conn.QueryContext(ctx2, q)
	if err != nil {
		return infer.FunctionResponse[GetStoresResult]{}, err
	}
	defer rows.Close()
	var list []GetStoreResult
	count := 0
	for rows.Next() {
		var name, typ, state, owner string
		var created, updated time.Time
		if err := rows.Scan(&name, &typ, &state, &owner, &created, &updated); err != nil {
			return infer.FunctionResponse[GetStoresResult]{}, err
		}
		count++
		if count > invokeMaxRows {
			break
		}
		list = append(list, GetStoreResult{Name: name, Type: typ, State: state, Owner: owner, CreatedAt: created.Format(time.RFC3339), UpdatedAt: updated.Format(time.RFC3339)})
	}
	if err := rows.Err(); err != nil {
		return infer.FunctionResponse[GetStoresResult]{}, err
	}
	if count > invokeMaxRows {
		log.Printf("pulumi-deltastream warning: getStores truncated at %d rows (more available)", invokeMaxRows)
	}
	return infer.FunctionResponse[GetStoresResult]{Output: GetStoresResult{Stores: list}}, nil
}
