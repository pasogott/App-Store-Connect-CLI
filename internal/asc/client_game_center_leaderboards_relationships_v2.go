package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetGameCenterLeaderboardSetMembersRelationshipsV2 retrieves leaderboard linkages for a v2 leaderboard set.
func (c *Client) GetGameCenterLeaderboardSetMembersRelationshipsV2(ctx context.Context, setID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getGameCenterLeaderboardSetLinkagesV2(ctx, setID, "gameCenterLeaderboards", opts...)
}

// AddGameCenterLeaderboardSetMembersV2 adds leaderboards to a v2 leaderboard set.
func (c *Client) AddGameCenterLeaderboardSetMembersV2(ctx context.Context, setID string, leaderboardIDs []string) error {
	payload := RelationshipRequest{
		Data: buildRelationshipData(ResourceTypeGameCenterLeaderboards, leaderboardIDs),
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v2/gameCenterLeaderboardSets/%s/relationships/gameCenterLeaderboards", strings.TrimSpace(setID))
	_, err = c.do(ctx, http.MethodPost, path, body)
	return err
}

// RemoveGameCenterLeaderboardSetMembersV2 removes leaderboards from a v2 leaderboard set.
func (c *Client) RemoveGameCenterLeaderboardSetMembersV2(ctx context.Context, setID string, leaderboardIDs []string) error {
	payload := RelationshipRequest{
		Data: buildRelationshipData(ResourceTypeGameCenterLeaderboards, leaderboardIDs),
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v2/gameCenterLeaderboardSets/%s/relationships/gameCenterLeaderboards", strings.TrimSpace(setID))
	_, err = c.do(ctx, http.MethodDelete, path, body)
	return err
}

// GetGameCenterLeaderboardSetVersionsRelationships retrieves version linkages for a v2 leaderboard set.
func (c *Client) GetGameCenterLeaderboardSetVersionsRelationships(ctx context.Context, setID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getGameCenterLeaderboardSetLinkagesV2(ctx, setID, "versions", opts...)
}

// GetGameCenterLeaderboardVersionsRelationships retrieves version linkages for a v2 leaderboard.
func (c *Client) GetGameCenterLeaderboardVersionsRelationships(ctx context.Context, leaderboardID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getGameCenterLeaderboardLinkagesV2(ctx, leaderboardID, "versions", opts...)
}

// UpdateGameCenterLeaderboardActivityRelationshipV2 updates the activity relationship on a v2 leaderboard.
func (c *Client) UpdateGameCenterLeaderboardActivityRelationshipV2(ctx context.Context, leaderboardID, activityID string) error {
	leaderboardID = strings.TrimSpace(leaderboardID)
	activityID = strings.TrimSpace(activityID)
	if leaderboardID == "" {
		return fmt.Errorf("leaderboardID is required")
	}
	if activityID == "" {
		return fmt.Errorf("activityID is required")
	}

	payload := struct {
		Data ResourceData `json:"data"`
	}{
		Data: ResourceData{
			Type: ResourceTypeGameCenterActivities,
			ID:   activityID,
		},
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v2/gameCenterLeaderboards/%s/relationships/activity", leaderboardID)
	_, err = c.do(ctx, http.MethodPatch, path, body)
	return err
}

// UpdateGameCenterLeaderboardChallengeRelationshipV2 updates the challenge relationship on a v2 leaderboard.
func (c *Client) UpdateGameCenterLeaderboardChallengeRelationshipV2(ctx context.Context, leaderboardID, challengeID string) error {
	leaderboardID = strings.TrimSpace(leaderboardID)
	challengeID = strings.TrimSpace(challengeID)
	if leaderboardID == "" {
		return fmt.Errorf("leaderboardID is required")
	}
	if challengeID == "" {
		return fmt.Errorf("challengeID is required")
	}

	payload := struct {
		Data ResourceData `json:"data"`
	}{
		Data: ResourceData{
			Type: ResourceTypeGameCenterChallenges,
			ID:   challengeID,
		},
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v2/gameCenterLeaderboards/%s/relationships/challenge", leaderboardID)
	_, err = c.do(ctx, http.MethodPatch, path, body)
	return err
}

func (c *Client) getGameCenterLeaderboardSetLinkagesV2(ctx context.Context, setID, relationship string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	setID = strings.TrimSpace(setID)
	if query.nextURL == "" && setID == "" {
		return nil, fmt.Errorf("setID is required")
	}

	path := fmt.Sprintf("/v2/gameCenterLeaderboardSets/%s/relationships/%s", setID, relationship)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("gameCenterLeaderboardSetRelationshipsV2: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response LinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *Client) getGameCenterLeaderboardLinkagesV2(ctx context.Context, leaderboardID, relationship string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	leaderboardID = strings.TrimSpace(leaderboardID)
	if query.nextURL == "" && leaderboardID == "" {
		return nil, fmt.Errorf("leaderboardID is required")
	}

	path := fmt.Sprintf("/v2/gameCenterLeaderboards/%s/relationships/%s", leaderboardID, relationship)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("gameCenterLeaderboardRelationshipsV2: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response LinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
