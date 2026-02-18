package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetBetaGroupBetaTestersRelationships retrieves beta tester linkages for a beta group.
func (c *Client) GetBetaGroupBetaTestersRelationships(ctx context.Context, groupID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getBetaGroupLinkages(ctx, groupID, "betaTesters", opts...)
}

// GetBetaGroupBuildsRelationships retrieves build linkages for a beta group.
func (c *Client) GetBetaGroupBuildsRelationships(ctx context.Context, groupID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getBetaGroupLinkages(ctx, groupID, "builds", opts...)
}

// BetaGroupAppLinkageResponse is the response for beta group app relationships.
type BetaGroupAppLinkageResponse struct {
	Data  ResourceData `json:"data"`
	Links Links        `json:"links"`
}

// BetaGroupBetaRecruitmentCriteriaLinkageResponse is the response for recruitment criteria relationships.
type BetaGroupBetaRecruitmentCriteriaLinkageResponse struct {
	Data  ResourceData `json:"data"`
	Links Links        `json:"links"`
}

// BetaGroupBetaRecruitmentCriterionCompatibleBuildCheckLinkageResponse is the response for compatible build check relationships.
type BetaGroupBetaRecruitmentCriterionCompatibleBuildCheckLinkageResponse struct {
	Data  ResourceData `json:"data"`
	Links Links        `json:"links"`
}

// GetBetaGroupAppRelationship retrieves the app linkage for a beta group.
func (c *Client) GetBetaGroupAppRelationship(ctx context.Context, groupID string) (*BetaGroupAppLinkageResponse, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}

	path := fmt.Sprintf("/v1/betaGroups/%s/relationships/app", groupID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BetaGroupAppLinkageResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &response, nil
}

// GetBetaGroupBetaRecruitmentCriteriaRelationship retrieves the recruitment criteria linkage for a beta group.
func (c *Client) GetBetaGroupBetaRecruitmentCriteriaRelationship(ctx context.Context, groupID string) (*BetaGroupBetaRecruitmentCriteriaLinkageResponse, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}

	path := fmt.Sprintf("/v1/betaGroups/%s/relationships/betaRecruitmentCriteria", groupID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BetaGroupBetaRecruitmentCriteriaLinkageResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &response, nil
}

// GetBetaGroupBetaRecruitmentCriterionCompatibleBuildCheckRelationship retrieves compatible build check linkage for a beta group.
func (c *Client) GetBetaGroupBetaRecruitmentCriterionCompatibleBuildCheckRelationship(ctx context.Context, groupID string) (*BetaGroupBetaRecruitmentCriterionCompatibleBuildCheckLinkageResponse, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}

	path := fmt.Sprintf("/v1/betaGroups/%s/relationships/betaRecruitmentCriterionCompatibleBuildCheck", groupID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BetaGroupBetaRecruitmentCriterionCompatibleBuildCheckLinkageResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &response, nil
}

// AddBuildsToBetaGroup adds builds to a beta group.
func (c *Client) AddBuildsToBetaGroup(ctx context.Context, groupID string, buildIDs []string) error {
	groupID = strings.TrimSpace(groupID)
	buildIDs = normalizeList(buildIDs)
	if groupID == "" {
		return fmt.Errorf("groupID is required")
	}
	if len(buildIDs) == 0 {
		return fmt.Errorf("buildIDs are required")
	}

	payload := RelationshipRequest{
		Data: make([]RelationshipData, len(buildIDs)),
	}
	for i, id := range buildIDs {
		payload.Data[i] = RelationshipData{
			Type: ResourceTypeBuilds,
			ID:   id,
		}
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/betaGroups/%s/relationships/builds", groupID)
	_, err = c.do(ctx, "POST", path, body)
	return err
}

// RemoveBuildsFromBetaGroup removes builds from a beta group.
func (c *Client) RemoveBuildsFromBetaGroup(ctx context.Context, groupID string, buildIDs []string) error {
	groupID = strings.TrimSpace(groupID)
	buildIDs = normalizeList(buildIDs)
	if groupID == "" {
		return fmt.Errorf("groupID is required")
	}
	if len(buildIDs) == 0 {
		return fmt.Errorf("buildIDs are required")
	}

	payload := RelationshipRequest{
		Data: make([]RelationshipData, len(buildIDs)),
	}
	for i, id := range buildIDs {
		payload.Data[i] = RelationshipData{
			Type: ResourceTypeBuilds,
			ID:   id,
		}
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/betaGroups/%s/relationships/builds", groupID)
	_, err = c.do(ctx, "DELETE", path, body)
	return err
}

// GetBetaTesterAppsRelationships retrieves app linkages for a beta tester.
func (c *Client) GetBetaTesterAppsRelationships(ctx context.Context, testerID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getBetaTesterLinkages(ctx, testerID, "apps", opts...)
}

// GetBetaTesterBetaGroupsRelationships retrieves beta group linkages for a beta tester.
func (c *Client) GetBetaTesterBetaGroupsRelationships(ctx context.Context, testerID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getBetaTesterLinkages(ctx, testerID, "betaGroups", opts...)
}

// GetBetaTesterBuildsRelationships retrieves build linkages for a beta tester.
func (c *Client) GetBetaTesterBuildsRelationships(ctx context.Context, testerID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getBetaTesterLinkages(ctx, testerID, "builds", opts...)
}

func (c *Client) getBetaGroupLinkages(ctx context.Context, groupID, relationship string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	groupID = strings.TrimSpace(groupID)
	if query.nextURL == "" && groupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}

	path := fmt.Sprintf("/v1/betaGroups/%s/relationships/%s", groupID, relationship)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("betaGroupRelationships: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response LinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *Client) getBetaTesterLinkages(ctx context.Context, testerID, relationship string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	testerID = strings.TrimSpace(testerID)
	if query.nextURL == "" && testerID == "" {
		return nil, fmt.Errorf("testerID is required")
	}

	path := fmt.Sprintf("/v1/betaTesters/%s/relationships/%s", testerID, relationship)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("betaTesterRelationships: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response LinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
