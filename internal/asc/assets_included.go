package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// includedSetMediaLimit is Apple's maximum for limit[appScreenshots] and
// limit[appPreviews]; a set holds at most 10 screenshots or 3 previews.
const includedSetMediaLimit = 50

func addIncludedSetMedia(values url.Values, relationship string) {
	values.Set("include", relationship)
	values.Set("limit["+relationship+"]", strconv.Itoa(includedSetMediaLimit))
}

// IncludedAppScreenshots maps set IDs to their included screenshots in the
// set's relationship order. Sets without complete linkage are omitted and need
// a per-set request.
func IncludedAppScreenshots(response *AppScreenshotSetsResponse) map[string][]Resource[AppScreenshotAttributes] {
	return includedSetMedia[AppScreenshotSetAttributes, AppScreenshotAttributes](response, "appScreenshots")
}

// IncludedAppPreviews maps set IDs to their included previews in the set's
// relationship order. Sets without complete linkage are omitted and need a
// per-set request.
func IncludedAppPreviews(response *AppPreviewSetsResponse) map[string][]Resource[AppPreviewAttributes] {
	return includedSetMedia[AppPreviewSetAttributes, AppPreviewAttributes](response, "appPreviews")
}

func includedSetMedia[S, T any](response *Response[S], relationship string) map[string][]Resource[T] {
	result := map[string][]Resource[T]{}
	if response == nil {
		return result
	}
	var included []json.RawMessage
	if len(response.Included) > 0 && json.Unmarshal(response.Included, &included) != nil {
		return result
	}
	byKey := make(map[ResourceData]json.RawMessage, len(included))
	for _, raw := range included {
		var key ResourceData
		if json.Unmarshal(raw, &key) == nil {
			byKey[key] = raw
		}
	}
	for _, set := range response.Data {
		var relationships map[string]struct {
			Data *[]ResourceData `json:"data"`
			Meta json.RawMessage `json:"meta"`
		}
		if json.Unmarshal(set.Relationships, &relationships) != nil {
			continue
		}
		linkage, ok := relationships[relationship]
		if !ok || linkage.Data == nil {
			continue
		}
		if total, ok := ParsePagingTotalOK(linkage.Meta); ok && total > len(*linkage.Data) {
			continue
		}
		items := make([]Resource[T], 0, len(*linkage.Data))
		for _, ref := range *linkage.Data {
			var item Resource[T]
			raw, ok := byKey[ref]
			if !ok || json.Unmarshal(raw, &item) != nil {
				break
			}
			items = append(items, item)
		}
		if len(items) == len(*linkage.Data) {
			result[set.ID] = items
		}
	}
	return result
}

// withoutLinkage drops the data and meta an include request adds to a set
// relationship, so set output matches a plain set read.
func withoutLinkage(relationships json.RawMessage, relationship string) json.RawMessage {
	var outer map[string]json.RawMessage
	if json.Unmarshal(relationships, &outer) != nil || outer[relationship] == nil {
		return relationships
	}
	var inner map[string]json.RawMessage
	if json.Unmarshal(outer[relationship], &inner) != nil {
		return relationships
	}
	if inner["data"] == nil && inner["meta"] == nil {
		return relationships
	}
	delete(inner, "data")
	delete(inner, "meta")
	encoded, err := json.Marshal(inner)
	if err != nil {
		return relationships
	}
	outer[relationship] = encoded
	encoded, err = json.Marshal(outer)
	if err != nil {
		return relationships
	}
	return encoded
}

// AppScreenshotSetsWithScreenshots pairs each set with its screenshots, using
// screenshots included with the sets and fetching only sets without complete
// linkage.
func (c *Client) AppScreenshotSetsWithScreenshots(ctx context.Context, response *AppScreenshotSetsResponse, requestContext RequestContextFunc) ([]AppScreenshotSetWithScreenshots, error) {
	included := IncludedAppScreenshots(response)
	sets := make([]AppScreenshotSetWithScreenshots, 0, len(response.Data))
	for _, set := range response.Data {
		screenshots, ok := included[set.ID]
		if !ok {
			fetched, err := c.GetAllAppScreenshots(ctx, set.ID, WithAppScreenshotsRequestContext(requestContext))
			if err != nil {
				return nil, fmt.Errorf("failed to fetch screenshots for set %s: %w", set.ID, err)
			}
			screenshots = fetched.Data
		}
		set.Relationships = withoutLinkage(set.Relationships, "appScreenshots")
		sets = append(sets, AppScreenshotSetWithScreenshots{Set: set, Screenshots: screenshots})
	}
	return sets, nil
}

// AppPreviewSetsWithPreviews pairs each set with its previews, using previews
// included with the sets and fetching only sets without complete linkage.
func (c *Client) AppPreviewSetsWithPreviews(ctx context.Context, response *AppPreviewSetsResponse, requestContext RequestContextFunc) ([]AppPreviewSetWithPreviews, error) {
	included := IncludedAppPreviews(response)
	sets := make([]AppPreviewSetWithPreviews, 0, len(response.Data))
	for _, set := range response.Data {
		previews, ok := included[set.ID]
		if !ok {
			requestCtx, cancel := requestContextFor(ctx, requestContext)
			fetched, err := c.GetAppPreviews(requestCtx, set.ID)
			cancel()
			if err != nil {
				return nil, fmt.Errorf("failed to fetch previews for set %s: %w", set.ID, err)
			}
			previews = fetched.Data
		}
		set.Relationships = withoutLinkage(set.Relationships, "appPreviews")
		sets = append(sets, AppPreviewSetWithPreviews{Set: set, Previews: previews})
	}
	return sets, nil
}
