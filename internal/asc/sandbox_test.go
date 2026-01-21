package asc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestBuildSandboxTestersQuery(t *testing.T) {
	query := &sandboxTestersQuery{}
	opts := []SandboxTestersOption{
		WithSandboxTestersEmail(" tester@example.com "),
		WithSandboxTestersTerritory("usa"),
		WithSandboxTestersLimit(10),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildSandboxTestersQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "10" {
		t.Fatalf("expected limit=10, got %q", got)
	}
}

func TestGetSandboxTesters_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"sandboxTesters","id":"1","attributes":{"acAccountName":"tester@example.com","firstName":"Test","lastName":"User","territory":"USA"}},{"type":"sandboxTesters","id":"2","attributes":{"acAccountName":"other@example.com","firstName":"Other","lastName":"User","territory":"JPN"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v2/sandboxTesters" {
			t.Fatalf("expected path /v2/sandboxTesters, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	resp, err := client.GetSandboxTesters(context.Background(),
		WithSandboxTestersEmail("tester@example.com"),
		WithSandboxTestersTerritory("usa"),
		WithSandboxTestersLimit(10),
	)
	if err != nil {
		t.Fatalf("GetSandboxTesters() error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 sandbox tester, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "1" {
		t.Fatalf("expected tester ID 1, got %q", resp.Data[0].ID)
	}
}

func TestGetSandboxTesters_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/sandboxTesters?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetSandboxTesters(context.Background(), WithSandboxTestersNextURL(next)); err != nil {
		t.Fatalf("GetSandboxTesters() error: %v", err)
	}
}

func TestGetSandboxTester_ByID(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"sandboxTesters","id":"tester-1","attributes":{"acAccountName":"tester@example.com","territory":"USA"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v2/sandboxTesters" {
			t.Fatalf("expected path /v2/sandboxTesters, got %s", req.URL.Path)
		}
		if got := req.URL.Query().Get("limit"); got != "200" {
			t.Fatalf("expected limit=200, got %q", got)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetSandboxTester(context.Background(), "tester-1"); err != nil {
		t.Fatalf("GetSandboxTester() error: %v", err)
	}
}

func TestCreateSandboxTester_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"sandboxTesters","id":"tester-1","attributes":{"email":"tester@example.com","firstName":"Test","lastName":"User","appStoreTerritory":"USA"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/sandboxTesters" {
			t.Fatalf("expected path /v1/sandboxTesters, got %s", req.URL.Path)
		}

		var body map[string]map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		data, ok := body["data"]
		if !ok {
			t.Fatalf("expected data in request body")
		}
		if data["type"] != "sandboxTesters" {
			t.Fatalf("expected type sandboxTesters, got %v", data["type"])
		}

		attributes, ok := data["attributes"].(map[string]any)
		if !ok {
			t.Fatalf("expected attributes in request body")
		}
		if attributes["email"] != "tester@example.com" {
			t.Fatalf("expected email tester@example.com, got %v", attributes["email"])
		}
		if attributes["firstName"] != "Test" || attributes["lastName"] != "User" {
			t.Fatalf("expected name Test User, got %v %v", attributes["firstName"], attributes["lastName"])
		}
		if attributes["appStoreTerritory"] != "USA" {
			t.Fatalf("expected appStoreTerritory USA, got %v", attributes["appStoreTerritory"])
		}
		assertAuthorized(t, req)
	}, response)

	attrs := SandboxTesterCreateAttributes{
		FirstName:         "Test",
		LastName:          "User",
		Email:             "tester@example.com",
		Password:          "Passwordtest1",
		ConfirmPassword:   "Passwordtest1",
		SecretQuestion:    "Question",
		SecretAnswer:      "Answer",
		BirthDate:         "1980-03-01",
		AppStoreTerritory: "USA",
	}
	if _, err := client.CreateSandboxTester(context.Background(), attrs); err != nil {
		t.Fatalf("CreateSandboxTester() error: %v", err)
	}
}

func TestDeleteSandboxTester(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/sandboxTesters/tester-1" {
			t.Fatalf("expected path /v1/sandboxTesters/tester-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteSandboxTester(context.Background(), "tester-1"); err != nil {
		t.Fatalf("DeleteSandboxTester() error: %v", err)
	}
}
