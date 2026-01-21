package cmd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

func TestFetchOptionalBuild_NotFound(t *testing.T) {
	resp, err := fetchOptionalBuild(context.Background(), "VERSION_ID", func(ctx context.Context, versionID string) (*asc.BuildResponse, error) {
		return nil, fmt.Errorf("NOT_FOUND: missing")
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
}

func TestFetchOptionalBuild_Error(t *testing.T) {
	expected := errors.New("boom")
	_, err := fetchOptionalBuild(context.Background(), "VERSION_ID", func(ctx context.Context, versionID string) (*asc.BuildResponse, error) {
		return nil, expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}

func TestFetchOptionalBuild_Success(t *testing.T) {
	resp, err := fetchOptionalBuild(context.Background(), "VERSION_ID", func(ctx context.Context, versionID string) (*asc.BuildResponse, error) {
		return &asc.BuildResponse{
			Data: asc.Resource[asc.BuildAttributes]{
				ID: "BUILD_ID",
				Attributes: asc.BuildAttributes{
					Version: "1.0",
				},
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil || resp.Data.ID != "BUILD_ID" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestFetchOptionalSubmission_NotFound(t *testing.T) {
	resp, err := fetchOptionalSubmission(context.Background(), "VERSION_ID", func(ctx context.Context, versionID string) (*asc.AppStoreVersionSubmissionResourceResponse, error) {
		return nil, fmt.Errorf("NOT_FOUND: missing")
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
}

func TestFetchOptionalSubmission_Error(t *testing.T) {
	expected := errors.New("boom")
	_, err := fetchOptionalSubmission(context.Background(), "VERSION_ID", func(ctx context.Context, versionID string) (*asc.AppStoreVersionSubmissionResourceResponse, error) {
		return nil, expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}

func TestFetchOptionalSubmission_Success(t *testing.T) {
	resp, err := fetchOptionalSubmission(context.Background(), "VERSION_ID", func(ctx context.Context, versionID string) (*asc.AppStoreVersionSubmissionResourceResponse, error) {
		return &asc.AppStoreVersionSubmissionResourceResponse{
			Data: asc.AppStoreVersionSubmissionResource{
				Type: asc.ResourceTypeAppStoreVersionSubmissions,
				ID:   "SUBMIT_ID",
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil || resp.Data.ID != "SUBMIT_ID" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
