package token_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/influxdata/influxdb/v2"
	ihttp "github.com/influxdata/influxdb/v2/http"
	"github.com/influxdata/influxdb/v2/inmem"
	"github.com/influxdata/influxdb/v2/kv"
	"github.com/influxdata/influxdb/v2/tenant"
	itesting "github.com/influxdata/influxdb/v2/testing"
	"github.com/influxdata/influxdb/v2/token"
	"go.uber.org/zap/zaptest"
)

func initAuthorizationService(f itesting.AuthorizationFields, t *testing.T) (influxdb.AuthorizationService, string, func()) {
	t.Helper()

	s, stCloser, err := NewTestInmemStore(t)
	if err != nil {
		t.Fatal(err)
	}

	storage, err := token.NewStore(s)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	svc := token.NewService(storage)

	// set up tenant service
	s, err := tenant.NewStore(s)
	if err != nil {
		t.Fatal(err)
	}

	ts := tenant.NewService(s)

	for _, u := range f.Users {
		if err := ts.CreateUser(ctx, u); err != nil {
			t.Fatalf("failed to populate users")
		}
	}

	for _, o := range f.Orgs {
		if err := ts.CreateOrganization(ctx, o); err != nil {
			t.Fatalf("failed to populate orgs")
		}
	}

	for _, u := range f.Authorizations {
		if err := svc.CreateAuthorization(ctx, u); err != nil {
			t.Fatalf("failed to populate authorizations")
		}
	}

	handler := token.NewHTTPAuthHandler(zaptest.NewLogger(t), svc), tenant
	r := chi.NewRouter()
	r.Mount(handler.Prefix(), handler)
	server := httptest.NewServer(r)
	httpClient, err := ihttp.NewHTTPClient(server.URL, "", false)
	if err != nil {
		t.Fatal(err)
	}

	client := token.AuthorizationClientService{
		Client: httpClient,
	}

	return &client, "http_token", func() {
		server.Close()
		stCloser()
	}

}

func NewTestInmemStore(t *testing.T) (kv.Store, func(), error) {
	return inmem.NewKVStore(), func() {}, nil
}

func TestAuthorizationService(t *testing.T) {
	itesting.AuthorizationService(initAuthorizationService, t)
}
