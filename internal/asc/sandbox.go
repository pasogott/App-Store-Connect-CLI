package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// SandboxTesterAttributes describes a sandbox tester resource.
type SandboxTesterAttributes struct {
	FirstName               string `json:"firstName,omitempty"`
	LastName                string `json:"lastName,omitempty"`
	Email                   string `json:"email,omitempty"`
	AccountName             string `json:"acAccountName,omitempty"`
	Password                string `json:"password,omitempty"`
	ConfirmPassword         string `json:"confirmPassword,omitempty"`
	SecretQuestion          string `json:"secretQuestion,omitempty"`
	SecretAnswer            string `json:"secretAnswer,omitempty"`
	BirthDate               string `json:"birthDate,omitempty"`
	AppStoreTerritory       string `json:"appStoreTerritory,omitempty"`
	Territory               string `json:"territory,omitempty"`
	ApplePayCompatible      *bool  `json:"applePayCompatible,omitempty"`
	InterruptPurchases      *bool  `json:"interruptPurchases,omitempty"`
	SubscriptionRenewalRate string `json:"subscriptionRenewalRate,omitempty"`
}

// SandboxTestersResponse is the response from sandbox testers endpoints.
type SandboxTestersResponse = Response[SandboxTesterAttributes]

// SandboxTesterResponse is the response from sandbox tester detail/creates.
type SandboxTesterResponse = SingleResponse[SandboxTesterAttributes]

// SandboxTesterCreateAttributes describes attributes for creating a sandbox tester.
type SandboxTesterCreateAttributes struct {
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	ConfirmPassword   string `json:"confirmPassword"`
	SecretQuestion    string `json:"secretQuestion"`
	SecretAnswer      string `json:"secretAnswer"`
	BirthDate         string `json:"birthDate"`
	AppStoreTerritory string `json:"appStoreTerritory"`
}

// SandboxTesterCreateData is the data portion of a sandbox tester create request.
type SandboxTesterCreateData struct {
	Type       ResourceType                  `json:"type"`
	Attributes SandboxTesterCreateAttributes `json:"attributes"`
}

// SandboxTesterCreateRequest is a request to create a sandbox tester.
type SandboxTesterCreateRequest struct {
	Data SandboxTesterCreateData `json:"data"`
}

type sandboxTestersQuery struct {
	listQuery
	email     string
	territory string
}

// SandboxTestersOption is a functional option for GetSandboxTesters.
type SandboxTestersOption func(*sandboxTestersQuery)

// WithSandboxTestersLimit sets the max number of sandbox testers to return.
func WithSandboxTestersLimit(limit int) SandboxTestersOption {
	return func(q *sandboxTestersQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithSandboxTestersNextURL uses a next page URL directly.
func WithSandboxTestersNextURL(next string) SandboxTestersOption {
	return func(q *sandboxTestersQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithSandboxTestersEmail filters sandbox testers by email address.
func WithSandboxTestersEmail(email string) SandboxTestersOption {
	return func(q *sandboxTestersQuery) {
		if strings.TrimSpace(email) != "" {
			q.email = strings.TrimSpace(email)
		}
	}
}

// WithSandboxTestersTerritory filters sandbox testers by App Store territory code (e.g., USA).
func WithSandboxTestersTerritory(territory string) SandboxTestersOption {
	return func(q *sandboxTestersQuery) {
		if strings.TrimSpace(territory) != "" {
			q.territory = strings.ToUpper(strings.TrimSpace(territory))
		}
	}
}

func buildSandboxTestersQuery(query *sandboxTestersQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func sandboxTesterEmail(attr SandboxTesterAttributes) string {
	if strings.TrimSpace(attr.Email) != "" {
		return strings.TrimSpace(attr.Email)
	}
	if strings.TrimSpace(attr.AccountName) != "" {
		return strings.TrimSpace(attr.AccountName)
	}
	return ""
}

func sandboxTesterTerritory(attr SandboxTesterAttributes) string {
	if strings.TrimSpace(attr.AppStoreTerritory) != "" {
		return strings.ToUpper(strings.TrimSpace(attr.AppStoreTerritory))
	}
	if strings.TrimSpace(attr.Territory) != "" {
		return strings.ToUpper(strings.TrimSpace(attr.Territory))
	}
	return ""
}

func filterSandboxTesters(items []Resource[SandboxTesterAttributes], email, territory string) []Resource[SandboxTesterAttributes] {
	normalizedEmail := strings.TrimSpace(email)
	normalizedTerritory := strings.ToUpper(strings.TrimSpace(territory))
	if normalizedEmail == "" && normalizedTerritory == "" {
		return items
	}
	filtered := make([]Resource[SandboxTesterAttributes], 0, len(items))
	for _, item := range items {
		if normalizedEmail != "" && !strings.EqualFold(sandboxTesterEmail(item.Attributes), normalizedEmail) {
			continue
		}
		if normalizedTerritory != "" && sandboxTesterTerritory(item.Attributes) != normalizedTerritory {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

// GetSandboxTesters retrieves sandbox testers with optional filters.
func (c *Client) GetSandboxTesters(ctx context.Context, opts ...SandboxTestersOption) (*SandboxTestersResponse, error) {
	query := &sandboxTestersQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v2/sandboxTesters"
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildSandboxTestersQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response SandboxTestersResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sandbox testers response: %w", err)
	}

	response.Data = filterSandboxTesters(response.Data, query.email, query.territory)
	return &response, nil
}

// GetSandboxTester retrieves a sandbox tester by ID.
func (c *Client) GetSandboxTester(ctx context.Context, testerID string) (*SandboxTesterResponse, error) {
	next := ""
	for {
		resp, err := c.GetSandboxTesters(ctx,
			WithSandboxTestersLimit(200),
			WithSandboxTestersNextURL(next),
		)
		if err != nil {
			return nil, err
		}
		for _, item := range resp.Data {
			if item.ID == testerID {
				return &SandboxTesterResponse{Data: item, Links: resp.Links}, nil
			}
		}
		if strings.TrimSpace(resp.Links.Next) == "" {
			break
		}
		next = resp.Links.Next
	}
	return nil, fmt.Errorf("sandbox tester not found: %s", testerID)
}

// CreateSandboxTester creates a sandbox tester.
func (c *Client) CreateSandboxTester(ctx context.Context, attributes SandboxTesterCreateAttributes) (*SandboxTesterResponse, error) {
	payload := SandboxTesterCreateRequest{
		Data: SandboxTesterCreateData{
			Type:       ResourceTypeSandboxTesters,
			Attributes: attributes,
		},
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/sandboxTesters", body)
	if err != nil {
		return nil, err
	}

	var response SandboxTesterResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sandbox tester response: %w", err)
	}

	return &response, nil
}

// DeleteSandboxTester deletes a sandbox tester by ID.
func (c *Client) DeleteSandboxTester(ctx context.Context, testerID string) error {
	path := fmt.Sprintf("/v1/sandboxTesters/%s", testerID)
	_, err := c.do(ctx, "DELETE", path, nil)
	return err
}
