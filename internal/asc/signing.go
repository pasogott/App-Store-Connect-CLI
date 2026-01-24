package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// BundleIDAttributes describes a bundle ID resource.
type BundleIDAttributes struct {
	Identifier string   `json:"identifier"`
	Name       string   `json:"name,omitempty"`
	Platform   Platform `json:"platform,omitempty"`
	SeedID     string   `json:"seedId,omitempty"`
}

// BundleIDsResponse is the response from bundle IDs endpoint.
type BundleIDsResponse = Response[BundleIDAttributes]

// BundleIDResponse is the response from bundle ID detail endpoint.
type BundleIDResponse = SingleResponse[BundleIDAttributes]

// CertificateAttributes describes a certificate resource.
type CertificateAttributes struct {
	CertificateType    string `json:"certificateType,omitempty"`
	CertificateContent string `json:"certificateContent,omitempty"`
	DisplayName        string `json:"displayName,omitempty"`
	ExpirationDate     string `json:"expirationDate,omitempty"`
	Name               string `json:"name,omitempty"`
	SerialNumber       string `json:"serialNumber,omitempty"`
}

// CertificatesResponse is the response from certificates endpoint.
type CertificatesResponse = Response[CertificateAttributes]

// CertificateResponse is the response from certificate detail endpoint.
type CertificateResponse = SingleResponse[CertificateAttributes]

// DeviceAttributes describes a device resource.
type DeviceAttributes struct {
	Name        string   `json:"name,omitempty"`
	DeviceClass string   `json:"deviceClass,omitempty"`
	Model       string   `json:"model,omitempty"`
	Platform    Platform `json:"platform,omitempty"`
	UDID        string   `json:"udid,omitempty"`
	Status      string   `json:"status,omitempty"`
	AddedDate   string   `json:"addedDate,omitempty"`
}

// DevicesResponse is the response from devices endpoint.
type DevicesResponse = Response[DeviceAttributes]

// ProfileState represents profile state values.
type ProfileState string

const (
	ProfileStateActive ProfileState = "ACTIVE"
)

// ProfileAttributes describes a profile resource.
type ProfileAttributes struct {
	Name           string       `json:"name,omitempty"`
	ProfileState   ProfileState `json:"profileState,omitempty"`
	ProfileType    string       `json:"profileType,omitempty"`
	ProfileContent string       `json:"profileContent,omitempty"`
	UUID           string       `json:"uuid,omitempty"`
	CreatedDate    string       `json:"createdDate,omitempty"`
	ExpirationDate string       `json:"expirationDate,omitempty"`
}

// ProfilesResponse is the response from profiles endpoint.
type ProfilesResponse = Response[ProfileAttributes]

// ProfileResponse is the response from profile detail endpoint.
type ProfileResponse = SingleResponse[ProfileAttributes]

type bundleIDsQuery struct {
	listQuery
	identifier string
}

// BundleIDsOption is a functional option for GetBundleIDs.
type BundleIDsOption func(*bundleIDsQuery)

// WithBundleIDsFilterIdentifier filters bundle IDs by identifier (supports CSV).
func WithBundleIDsFilterIdentifier(identifier string) BundleIDsOption {
	return func(q *bundleIDsQuery) {
		normalized := normalizeCSVString(identifier)
		if normalized != "" {
			q.identifier = normalized
		}
	}
}

// WithBundleIDsLimit sets the max number of bundle IDs to return.
func WithBundleIDsLimit(limit int) BundleIDsOption {
	return func(q *bundleIDsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBundleIDsNextURL uses a next page URL directly.
func WithBundleIDsNextURL(next string) BundleIDsOption {
	return func(q *bundleIDsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

type certificatesQuery struct {
	listQuery
	certificateType string
}

// CertificatesOption is a functional option for GetCertificates.
type CertificatesOption func(*certificatesQuery)

// WithCertificatesFilterType filters certificates by certificate type (supports CSV).
func WithCertificatesFilterType(certType string) CertificatesOption {
	return func(q *certificatesQuery) {
		normalized := normalizeUpperCSVString(certType)
		if normalized != "" {
			q.certificateType = normalized
		}
	}
}

// WithCertificatesLimit sets the max number of certificates to return.
func WithCertificatesLimit(limit int) CertificatesOption {
	return func(q *certificatesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithCertificatesNextURL uses a next page URL directly.
func WithCertificatesNextURL(next string) CertificatesOption {
	return func(q *certificatesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

type profilesQuery struct {
	listQuery
	bundleID    string
	profileType string
}

// ProfilesOption is a functional option for GetProfiles.
type ProfilesOption func(*profilesQuery)

// WithProfilesFilterBundleID filters profiles by bundle ID.
func WithProfilesFilterBundleID(bundleID string) ProfilesOption {
	return func(q *profilesQuery) {
		if strings.TrimSpace(bundleID) != "" {
			q.bundleID = strings.TrimSpace(bundleID)
		}
	}
}

// WithProfilesFilterType filters profiles by profile type.
func WithProfilesFilterType(profileType string) ProfilesOption {
	return func(q *profilesQuery) {
		normalized := normalizeUpperCSVString(profileType)
		if normalized != "" {
			q.profileType = normalized
		}
	}
}

// WithProfilesLimit sets the max number of profiles to return.
func WithProfilesLimit(limit int) ProfilesOption {
	return func(q *profilesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithProfilesNextURL uses a next page URL directly.
func WithProfilesNextURL(next string) ProfilesOption {
	return func(q *profilesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

type devicesQuery struct {
	listQuery
	udids     []string
	platforms []string
	statuses  []string
}

// DevicesOption is a functional option for GetDevices.
type DevicesOption func(*devicesQuery)

// WithDevicesFilterUDIDs filters devices by UDID(s).
func WithDevicesFilterUDIDs(udids []string) DevicesOption {
	return func(q *devicesQuery) {
		q.udids = normalizeList(udids)
	}
}

// WithDevicesFilterPlatforms filters devices by platform(s).
func WithDevicesFilterPlatforms(platforms []string) DevicesOption {
	return func(q *devicesQuery) {
		q.platforms = normalizeUpperList(platforms)
	}
}

// WithDevicesFilterStatuses filters devices by status (e.g., ENABLED, DISABLED).
func WithDevicesFilterStatuses(statuses []string) DevicesOption {
	return func(q *devicesQuery) {
		q.statuses = normalizeUpperList(statuses)
	}
}

// WithDevicesLimit sets the max number of devices to return.
func WithDevicesLimit(limit int) DevicesOption {
	return func(q *devicesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithDevicesNextURL uses a next page URL directly.
func WithDevicesNextURL(next string) DevicesOption {
	return func(q *devicesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// ProfileCreateAttributes describes attributes for creating a profile.
type ProfileCreateAttributes struct {
	Name           string   `json:"name"`
	ProfileType    string   `json:"profileType"`
	BundleIDID     string   `json:"-"`
	CertificateIDs []string `json:"-"`
	DeviceIDs      []string `json:"-"`
}

// ProfileRelationships describes relationships for profiles.
type ProfileRelationships struct {
	BundleID     *Relationship     `json:"bundleId"`
	Certificates *RelationshipList `json:"certificates,omitempty"`
	Devices      *RelationshipList `json:"devices,omitempty"`
}

// ProfileCreateData is the data portion of a profile create request.
type ProfileCreateData struct {
	Type          ResourceType            `json:"type"`
	Attributes    ProfileCreateAttributes `json:"attributes"`
	Relationships *ProfileRelationships   `json:"relationships"`
}

// ProfileCreateRequest is a request to create a profile.
type ProfileCreateRequest struct {
	Data ProfileCreateData `json:"data"`
}

func buildBundleIDsQuery(query *bundleIDsQuery) string {
	values := url.Values{}
	if strings.TrimSpace(query.identifier) != "" {
		values.Set("filter[identifier]", strings.TrimSpace(query.identifier))
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildCertificatesQuery(query *certificatesQuery) string {
	values := url.Values{}
	if strings.TrimSpace(query.certificateType) != "" {
		values.Set("filter[certificateType]", strings.TrimSpace(query.certificateType))
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildProfilesQuery(query *profilesQuery) string {
	values := url.Values{}
	if strings.TrimSpace(query.bundleID) != "" {
		values.Set("filter[bundleId]", strings.TrimSpace(query.bundleID))
	}
	if strings.TrimSpace(query.profileType) != "" {
		values.Set("filter[profileType]", strings.TrimSpace(query.profileType))
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildDevicesQuery(query *devicesQuery) string {
	values := url.Values{}
	addCSV(values, "filter[udid]", query.udids)
	addCSV(values, "filter[platform]", query.platforms)
	addCSV(values, "filter[status]", query.statuses)
	addLimit(values, query.limit)
	return values.Encode()
}

// GetBundleIDs retrieves bundle IDs with optional filters.
func (c *Client) GetBundleIDs(ctx context.Context, opts ...BundleIDsOption) (*BundleIDsResponse, error) {
	query := &bundleIDsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/bundleIds"
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("bundleIds: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildBundleIDsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BundleIDsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCertificates retrieves certificates with optional filters.
func (c *Client) GetCertificates(ctx context.Context, opts ...CertificatesOption) (*CertificatesResponse, error) {
	query := &certificatesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/certificates"
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("certificates: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildCertificatesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CertificatesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetProfiles retrieves profiles with optional filters.
func (c *Client) GetProfiles(ctx context.Context, opts ...ProfilesOption) (*ProfilesResponse, error) {
	query := &profilesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/profiles"
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("profiles: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildProfilesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response ProfilesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetDevices retrieves devices with optional filters.
func (c *Client) GetDevices(ctx context.Context, opts ...DevicesOption) (*DevicesResponse, error) {
	query := &devicesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/devices"
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("devices: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildDevicesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response DevicesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateProfile creates a provisioning profile.
func (c *Client) CreateProfile(ctx context.Context, attrs ProfileCreateAttributes) (*ProfileResponse, error) {
	certIDs := normalizeList(attrs.CertificateIDs)
	deviceIDs := normalizeList(attrs.DeviceIDs)

	relationships := &ProfileRelationships{
		BundleID: &Relationship{
			Data: ResourceData{
				Type: ResourceTypeBundleIDs,
				ID:   strings.TrimSpace(attrs.BundleIDID),
			},
		},
	}

	if len(certIDs) > 0 {
		data := make([]ResourceData, 0, len(certIDs))
		for _, id := range certIDs {
			data = append(data, ResourceData{Type: ResourceTypeCertificates, ID: id})
		}
		relationships.Certificates = &RelationshipList{Data: data}
	}

	if len(deviceIDs) > 0 {
		data := make([]ResourceData, 0, len(deviceIDs))
		for _, id := range deviceIDs {
			data = append(data, ResourceData{Type: ResourceTypeDevices, ID: id})
		}
		relationships.Devices = &RelationshipList{Data: data}
	}

	payload := ProfileCreateRequest{
		Data: ProfileCreateData{
			Type: ResourceTypeProfiles,
			Attributes: ProfileCreateAttributes{
				Name:        strings.TrimSpace(attrs.Name),
				ProfileType: strings.TrimSpace(attrs.ProfileType),
			},
			Relationships: relationships,
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/profiles", body)
	if err != nil {
		return nil, err
	}

	var response ProfileResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
