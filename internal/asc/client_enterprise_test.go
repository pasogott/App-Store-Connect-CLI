package asc

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

const (
	enterpriseCertificatesURL = EnterpriseBaseURL + "/v1/certificates"
	ascCertificatesURL        = BaseURL + "/v1/certificates"
)

func requestAudience(t *testing.T, req *http.Request) []string {
	t.Helper()
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	parsed, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("ParseUnverified() error: %v", err)
	}
	audience, err := parsed.Claims.GetAudience()
	if err != nil {
		t.Fatalf("GetAudience() error: %v", err)
	}
	return audience
}

func TestEnterpriseClientUsesEnterpriseHostAudienceAndNextLinks(t *testing.T) {
	var requested []string
	client := newTestClient(t, func(req *http.Request) {
		requested = append(requested, req.URL.String())
		audience := requestAudience(t, req)
		if len(audience) != 1 || audience[0] != "apple-developer-enterprise-v1" {
			t.Fatalf("audience = %v", audience)
		}
	},
		jsonResponse(http.StatusOK, `{"data":[{"type":"certificates","id":"1"}],"links":{"next":"`+enterpriseCertificatesURL+`?cursor=two"}}`),
		jsonResponse(http.StatusOK, `{"data":[{"type":"certificates","id":"2"}]}`),
	)
	client.UseEnterpriseProgramAPI()

	ctx := context.Background()
	first, err := client.GetCertificates(ctx)
	if err != nil {
		t.Fatalf("GetCertificates() error: %v", err)
	}
	all, err := PaginateAll(ctx, first, func(ctx context.Context, nextURL string) (PaginatedResponse, error) {
		return client.GetCertificates(ctx, WithCertificatesNextURL(nextURL))
	})
	if err != nil {
		t.Fatalf("PaginateAll() error: %v", err)
	}

	if got := len(all.(*CertificatesResponse).Data); got != 2 {
		t.Fatalf("certificates = %d, want 2", got)
	}
	want := []string{enterpriseCertificatesURL, enterpriseCertificatesURL + "?cursor=two"}
	if strings.Join(requested, " ") != strings.Join(want, " ") {
		t.Fatalf("requested = %v, want %v", requested, want)
	}
}

func TestClientRejectsAbsoluteURLOnTheOtherAPIHost(t *testing.T) {
	tests := []struct {
		name       string
		enterprise bool
		target     string
	}{
		{name: "app store connect client", target: enterpriseCertificatesURL + "?cursor=two"},
		{name: "enterprise client", enterprise: true, target: ascCertificatesURL + "?cursor=two"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newTestClient(t, func(req *http.Request) {
				t.Fatalf("unexpected request to %s", req.URL)
			}, jsonResponse(http.StatusOK, `{"data":[]}`))
			if test.enterprise {
				client.UseEnterpriseProgramAPI()
			}

			_, err := client.GetCertificates(context.Background(), WithCertificatesNextURL(test.target))
			if err == nil || !strings.Contains(err.Error(), "rejected request URL") {
				t.Fatalf("GetCertificates() error = %v", err)
			}
			if _, err := client.RawRequest(context.Background(), http.MethodGet, test.target, nil); err == nil || !strings.Contains(err.Error(), "rejected request URL") {
				t.Fatalf("RawRequest() error = %v", err)
			}
		})
	}
}

func TestPaginationURLIdentityTreatsEnterpriseLinksAsRelative(t *testing.T) {
	absolute := PaginationURLIdentity(enterpriseCertificatesURL + "?limit=10&cursor=two")
	relative := PaginationURLIdentity("/v1/certificates?cursor=two&limit=10")
	if absolute != relative {
		t.Fatalf("identity = %q, want %q", absolute, relative)
	}
}
