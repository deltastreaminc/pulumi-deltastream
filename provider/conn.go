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
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	p "github.com/pulumi/pulumi-go-provider"

	ds "github.com/deltastreaminc/go-deltastream"
)

// buildHTTPClient builds an HTTP client with TLS and timeouts similar to Terraform provider.
func buildHTTPClient(insecureSkipVerify bool, sessionID *string) *http.Client {
	tlsConfig := &tls.Config{}
	if insecureSkipVerify {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	tr := &http.Transport{
		Dial:                  (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 20 * time.Second}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 1 * time.Minute,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       5 * time.Minute,
		TLSClientConfig:       tlsConfig,
		DisableKeepAlives:     true,
		MaxIdleConnsPerHost:   -1,
	}

	// Add a simple UA with optional session id.
	rt := roundTripperWithUA{r: tr, sessionID: sessionID}
	return &http.Client{Transport: rt}
}

type roundTripperWithUA struct {
	r         http.RoundTripper
	sessionID *string
}

func (d roundTripperWithUA) RoundTrip(h *http.Request) (*http.Response, error) {
	ua := "pulumi-provider-deltastream"
	if d.sessionID != nil {
		ua += " session/" + *d.sessionID
	}
	h.Header.Set("User-Agent", ua)
	return d.r.RoundTrip(h)
}

// openDB returns an sql.DB configured with server, api key, and HTTP client.
func openDB(ctx context.Context, cfg *Config) (*sql.DB, error) {
	logger := p.GetLogger(ctx)

	if cfg.APIKey == nil || *cfg.APIKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}
	if cfg.Server == nil || *cfg.Server == "" {
		return nil, fmt.Errorf("server is required")
	}
	server := *cfg.Server

	httpClient := buildHTTPClient(cfg.InsecureSkipVerify != nil && *cfg.InsecureSkipVerify, cfg.SessionID)

	opts := []ds.ConnectionOption{ds.WithStaticToken(*cfg.APIKey), ds.WithServer(server), ds.WithHTTPClient(httpClient)}
	if cfg.SessionID != nil && *cfg.SessionID != "" {
		opts = append(opts, ds.WithSessionID(*cfg.SessionID))
	}

	connector, err := ds.ConnectorWithOptions(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}
	db := sql.OpenDB(connector)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping server: %w", err)
	}
	logger.Debug("DeltaStream connection initialized")
	return db, nil
}

// withOrgRole applies organization and role to the underlying driver connection context.
func withOrgRole(ctx context.Context, db *sql.DB, org, role string) (context.Context, *sql.Conn, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return ctx, nil, err
	}

	conn.Raw(func(driverConn interface{}) error {
		if c, ok := driverConn.(*ds.Conn); ok {
			rsctx := c.GetContext()
			if org != "" {
				if id, err := uuid.Parse(org); err == nil {
					rsctx.OrganizationID = &id
				}
			}
			if role != "" {
				rsctx.RoleName = &role
			}
			c.SetContext(rsctx)
		}
		return nil
	})
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return ctx, nil, fmt.Errorf("failed to establish connection: %w", err)
	}
	return ctx, conn, nil
}

// quoteIdent returns a double-quoted SQL identifier with embedded quotes escaped.
// It does not attempt full normalization beyond doubling internal quotes.
func quoteIdent(in string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(in, "\"", "\"\""))
}

func quoteString(in string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(in, "'", "''"))
}
