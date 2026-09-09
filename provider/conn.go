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
	"net/http/httputil"
	"os"
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
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 20 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 1 * time.Minute,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       5 * time.Minute,
		TLSClientConfig:       tlsConfig,
		DisableKeepAlives:     true,
		MaxIdleConnsPerHost:   -1,
	}

	// Add a simple UA with optional session id.
	var rt http.RoundTripper = roundTripperWithUA{r: tr, sessionID: sessionID}
	if logPath := os.Getenv("DELTASTREAM_PULUMI_HTTP_DEBUG"); logPath != "" {
		rt = &debugTransport{r: rt, logPath: logPath}
	}
	return &http.Client{Transport: rt}
}

// debugTransport dumps full raw HTTP requests/responses (headers + body) to a
// file, mirroring the CLI's debugTransport (deltastreamv2/client/lib/http-client.go)
// so we can inspect the exact bytes submitted by the Pulumi provider.
// Enabled only when DELTASTREAM_PULUMI_HTTP_DEBUG is set to a writable file path.
type debugTransport struct {
	r       http.RoundTripper
	logPath string
}

func (d *debugTransport) RoundTrip(h *http.Request) (*http.Response, error) {
	// The dumped request/response may contain secrets (e.g. API keys, auth
	// tokens), so restrict the log file to owner-only access.
	f, ferr := os.OpenFile(d.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if ferr != nil {
		// If we can't open the log file, skip the (potentially expensive)
		// dump/RoundTrip logging entirely and just forward the request.
		return d.r.RoundTrip(h)
	}
	defer f.Close() //nolint:errcheck

	logf := func(format string, args ...interface{}) {
		_, _ = fmt.Fprintf(f, format, args...)
	}

	dump, _ := httputil.DumpRequestOut(h, true)
	logf("===== REQUEST %s =====\n%s\n", time.Now().Format(time.RFC3339Nano), string(dump))

	resp, err := d.r.RoundTrip(h)
	if err != nil {
		logf("===== ROUNDTRIP ERROR =====\n%v\n\n", err)
		return resp, err
	}
	if resp != nil {
		dumpResp, _ := httputil.DumpResponse(resp, true)
		logf("===== RESPONSE %s =====\n%s\n\n", time.Now().Format(time.RFC3339Nano), string(dumpResp))
	} else {
		logf("===== RESPONSE is nil =====\n\n")
	}
	return resp, err
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

	if err := conn.Raw(func(driverConn interface{}) error {
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
	}); err != nil {
		_ = conn.Close()
		return ctx, nil, fmt.Errorf("failed to configure connection context: %w", err)
	}
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

// setStringIfPresent sets changes[key] to a quoted SQL string literal derived
// from valPtr, but only when valPtr is non-nil and non-empty. It is a no-op
// otherwise, leaving key absent from changes rather than sending NULL. This
// is intended for STRING_VALUE-typed UPDATE STORE parameters that do not
// accept a bare NULL literal (e.g. kafka.sasl.username, kafka.sasl.password,
// kafka.msk.iam_role_arn, kafka.msk.aws_region): when the field doesn't
// apply to the current configuration, omitting it lets the backend surface
// its own "required parameter" validation error if the field actually is
// needed.
func setStringIfPresent(changes map[string]string, key string, valPtr *string) {
	if valPtr == nil || *valPtr == "" {
		return
	}
	changes[key] = quoteString(*valPtr)
}
