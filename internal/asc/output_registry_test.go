package asc

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestOutputRegistryNotEmpty(t *testing.T) {
	if len(outputRegistry) == 0 {
		t.Fatal("output registry is empty; init() may not have run")
	}
}

func TestOutputRegistryAllHandlersNonNil(t *testing.T) {
	for typ, fn := range outputRegistry {
		if fn == nil {
			t.Errorf("nil handler registered for type %s", typ)
		}
	}
}

func TestOutputRegistryExpectedTypeCount(t *testing.T) {
	// Total registered types across both registries should be ~471.
	total := len(outputRegistry) + len(directRenderRegistry)
	const minExpected = 460
	if total < minExpected {
		t.Errorf("expected at least %d registered types, got %d (rows: %d, direct: %d)",
			minExpected, total, len(outputRegistry), len(directRenderRegistry))
	}
}

func TestDirectRenderRegistryAllHandlersNonNil(t *testing.T) {
	for typ, fn := range directRenderRegistry {
		if fn == nil {
			t.Errorf("nil handler registered for type %s", typ)
		}
	}
}

func TestRenderByRegistryFallbackToJSON(t *testing.T) {
	// Unregistered type should fall back to JSON without error.
	type unregistered struct {
		Value string `json:"value"`
	}
	output := captureStdout(t, func() error {
		return renderByRegistry(&unregistered{Value: "test"}, RenderTable)
	})
	if output == "" {
		t.Fatal("expected JSON fallback output")
	}
	if !strings.Contains(output, "test") {
		t.Fatalf("expected JSON output to contain 'test', got: %s", output)
	}
}

func TestOutputRegistrySingleLinkageHelperRegistration(t *testing.T) {
	handler := requireOutputHandler(
		t,
		reflect.TypeOf(&AppStoreVersionSubmissionLinkageResponse{}),
		"AppStoreVersionSubmissionLinkageResponse",
	)

	headers, rows, err := handler(&AppStoreVersionSubmissionLinkageResponse{
		Data: ResourceData{
			Type: ResourceType("appStoreVersionSubmissions"),
			ID:   "submission-123",
		},
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	assertRowContains(t, headers, rows, 2, "submission-123")
}

func TestOutputRegistrySingleLinkageHelperPanicsOnNilExtractor(t *testing.T) {
	type linkage struct{}

	key := reflect.TypeOf(&linkage{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil linkage extractor", func() {
		registerSingleLinkageRows[linkage](nil)
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistrySingleLinkageHelperNilExtractorPanicsBeforeConflictChecks(t *testing.T) {
	type linkage struct{}

	preregisterRowsForConflict[linkage](t, "id")

	expectPanicContains(t, "nil linkage extractor", func() {
		registerSingleLinkageRows[linkage](nil)
	})
}

func TestOutputRegistryIDStateHelperRegistration(t *testing.T) {
	handler := requireOutputHandler(
		t,
		reflect.TypeOf(&BackgroundAssetVersionAppStoreReleaseResponse{}),
		"BackgroundAssetVersionAppStoreReleaseResponse",
	)

	headers, rows, err := handler(&BackgroundAssetVersionAppStoreReleaseResponse{
		Data: Resource[BackgroundAssetVersionAppStoreReleaseAttributes]{
			ID:         "release-1",
			Attributes: BackgroundAssetVersionAppStoreReleaseAttributes{State: "READY_FOR_SALE"},
		},
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	assertRowContains(t, headers, rows, 2, "release-1", "READY_FOR_SALE")
}

func TestOutputRegistryIDStateHelperPanicsOnNilExtractor(t *testing.T) {
	type state struct{}

	key := reflect.TypeOf(&state{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil id/state extractor", func() {
		registerIDStateRows[state](nil, func(id, value string) ([]string, [][]string) {
			return []string{"id", "state"}, [][]string{{id, value}}
		})
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryIDStateHelperNilExtractorPanicsBeforeConflictChecks(t *testing.T) {
	type state struct{}

	preregisterRowsForConflict[state](t, "id", "state")

	expectPanicContains(t, "nil id/state extractor", func() {
		registerIDStateRows[state](nil, func(id, value string) ([]string, [][]string) {
			return []string{"id", "state"}, [][]string{{id, value}}
		})
	})
}

func TestOutputRegistryIDStateHelperPanicsOnNilRows(t *testing.T) {
	type state struct{}

	key := reflect.TypeOf(&state{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil id/state rows function", func() {
		registerIDStateRows[state](func(*state) (string, string) {
			return "id", "value"
		}, nil)
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryIDStateHelperNilRowsPanicsBeforeConflictChecks(t *testing.T) {
	type state struct{}

	preregisterRowsForConflict[state](t, "id", "state")

	expectPanicContains(t, "nil id/state rows function", func() {
		registerIDStateRows[state](func(*state) (string, string) {
			return "id", "value"
		}, nil)
	})
}

func TestOutputRegistryIDBoolHelperRegistration(t *testing.T) {
	handler := requireOutputHandler(
		t,
		reflect.TypeOf(&AlternativeDistributionDomainDeleteResult{}),
		"AlternativeDistributionDomainDeleteResult",
	)

	headers, rows, err := handler(&AlternativeDistributionDomainDeleteResult{
		ID:      "domain-1",
		Deleted: true,
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	assertRowContains(t, headers, rows, 2, "domain-1", "true")
}

func TestOutputRegistryIDBoolHelperPanicsOnNilRows(t *testing.T) {
	type idBool struct{}

	key := reflect.TypeOf(&idBool{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil id/bool rows function", func() {
		registerIDBoolRows[idBool](func(*idBool) (string, bool) {
			return "id", true
		}, nil)
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryIDBoolHelperPanicsOnNilExtractor(t *testing.T) {
	type idBool struct{}

	key := reflect.TypeOf(&idBool{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil id/bool extractor", func() {
		registerIDBoolRows[idBool](nil, func(id string, deleted bool) ([]string, [][]string) {
			return []string{"id", "deleted"}, [][]string{{id, fmt.Sprintf("%t", deleted)}}
		})
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryIDBoolHelperNilExtractorPanicsBeforeConflictChecks(t *testing.T) {
	type idBool struct{}

	preregisterRowsForConflict[idBool](t, "id", "deleted")

	expectPanicContains(t, "nil id/bool extractor", func() {
		registerIDBoolRows[idBool](nil, func(id string, deleted bool) ([]string, [][]string) {
			return []string{"id", "deleted"}, [][]string{{id, fmt.Sprintf("%t", deleted)}}
		})
	})
}

func TestOutputRegistryResponseDataHelperRegistration(t *testing.T) {
	handler := requireOutputHandler(
		t,
		reflect.TypeOf(&Response[BetaGroupMetricAttributes]{}),
		"Response[BetaGroupMetricAttributes]",
	)

	headers, rows, err := handler(&Response[BetaGroupMetricAttributes]{
		Data: []Resource[BetaGroupMetricAttributes]{
			{
				ID:         "metric-1",
				Attributes: BetaGroupMetricAttributes{"installs": 12},
			},
		},
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	assertRowContains(t, headers, rows, 2, "metric-1", "installs=12")
}

func TestOutputRegistryResponseDataHelperPanicsOnNilRows(t *testing.T) {
	type attrs struct{}

	key := reflect.TypeOf(&Response[attrs]{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil response-data rows function", func() {
		registerResponseDataRows[attrs](nil)
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryResponseDataHelperNilRowsPanicsBeforeConflictChecks(t *testing.T) {
	type attrs struct{}

	preregisterRowsForConflict[Response[attrs]](t, "id")

	expectPanicContains(t, "nil response-data rows function", func() {
		registerResponseDataRows[attrs](nil)
	})
}

func TestOutputRegistrySingleResourceHelperRegistration(t *testing.T) {
	type helperAttrs struct {
		Name string `json:"name"`
	}

	registerSingleResourceRowsAdapter(func(v *Response[helperAttrs]) ([]string, [][]string) {
		if len(v.Data) == 0 {
			return []string{"ID", "Name"}, nil
		}
		return []string{"ID", "Name"}, [][]string{{v.Data[0].ID, v.Data[0].Attributes.Name}}
	})

	key := reflect.TypeOf(&SingleResponse[helperAttrs]{})
	cleanupRegistryTypes(t, key)

	handler := requireOutputHandler(t, key, "SingleResponse helper")

	headers, rows, err := handler(&SingleResponse[helperAttrs]{
		Data: Resource[helperAttrs]{
			ID:         "helper-id",
			Attributes: helperAttrs{Name: "helper-name"},
		},
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if len(headers) != 2 || headers[0] != "ID" || headers[1] != "Name" {
		t.Fatalf("unexpected headers: %v", headers)
	}
	if len(rows) != 1 || len(rows[0]) != 2 {
		t.Fatalf("unexpected rows shape: %v", rows)
	}
	if rows[0][0] != "helper-id" || rows[0][1] != "helper-name" {
		t.Fatalf("unexpected row: %v", rows[0])
	}
}

func TestOutputRegistrySingleResourceHelperPanicsOnNilRowsFunction(t *testing.T) {
	type helperAttrs struct {
		Name string `json:"name"`
	}

	singleKey := reflect.TypeOf(&SingleResponse[helperAttrs]{})
	cleanupRegistryTypes(t, singleKey)

	expectPanicContains(t, "nil rows function", func() {
		registerSingleResourceRowsAdapter[helperAttrs](nil)
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistrySingleResourceHelperNilRowsPanicsBeforeConflictChecks(t *testing.T) {
	type helperAttrs struct {
		Name string `json:"name"`
	}

	preregisterRowsForConflict[SingleResponse[helperAttrs]](t, "ID")

	expectPanicContains(t, "nil rows function", func() {
		registerSingleResourceRowsAdapter[helperAttrs](nil)
	})
}

func TestOutputRegistryRowsWithSingleResourceHelperRegistration(t *testing.T) {
	type attrs struct {
		Name string `json:"name"`
	}

	registerRowsWithSingleResourceAdapter(func(v *Response[attrs]) ([]string, [][]string) {
		if len(v.Data) == 0 {
			return []string{"ID", "Name"}, nil
		}
		return []string{"ID", "Name"}, [][]string{{v.Data[0].ID, v.Data[0].Attributes.Name}}
	})

	listKey := reflect.TypeOf(&Response[attrs]{})
	singleKey := reflect.TypeOf(&SingleResponse[attrs]{})
	cleanupRegistryTypes(t, listKey, singleKey)

	listHandler := requireOutputHandler(t, listKey, "list handler from rows+single-resource helper")
	singleHandler := requireOutputHandler(t, singleKey, "single handler from rows+single-resource helper")

	_, listRows, err := listHandler(&Response[attrs]{
		Data: []Resource[attrs]{{ID: "list-id", Attributes: attrs{Name: "list-name"}}},
	})
	if err != nil {
		t.Fatalf("list handler returned error: %v", err)
	}
	if len(listRows) != 1 || len(listRows[0]) != 2 || listRows[0][0] != "list-id" || listRows[0][1] != "list-name" {
		t.Fatalf("unexpected list rows: %v", listRows)
	}

	_, singleRows, err := singleHandler(&SingleResponse[attrs]{
		Data: Resource[attrs]{ID: "single-id", Attributes: attrs{Name: "single-name"}},
	})
	if err != nil {
		t.Fatalf("single handler returned error: %v", err)
	}
	if len(singleRows) != 1 || len(singleRows[0]) != 2 || singleRows[0][0] != "single-id" || singleRows[0][1] != "single-name" {
		t.Fatalf("unexpected single rows: %v", singleRows)
	}
}

func TestOutputRegistryRowsWithSingleResourceHelperNoPartialRegistrationOnPanic(t *testing.T) {
	type attrs struct {
		Name string `json:"name"`
	}

	listKey := reflect.TypeOf(&Response[attrs]{})
	singleKey := reflect.TypeOf(&SingleResponse[attrs]{})
	cleanupRegistryTypes(t, listKey, singleKey)

	registerRows(func(v *SingleResponse[attrs]) ([]string, [][]string) {
		return []string{"ID"}, [][]string{{v.Data.ID}}
	})

	expectPanic(t, "expected conflict panic when single handler is already registered", func() {
		registerRowsWithSingleResourceAdapter(func(v *Response[attrs]) ([]string, [][]string) {
			return []string{"ID"}, nil
		})
	})

	assertRegistryTypeAbsent(t, listKey)
}

func TestOutputRegistryRowsWithSingleResourceHelperNoPartialRegistrationWhenListRegistered(t *testing.T) {
	type attrs struct {
		Name string `json:"name"`
	}

	listKey := reflect.TypeOf(&Response[attrs]{})
	singleKey := reflect.TypeOf(&SingleResponse[attrs]{})
	cleanupRegistryTypes(t, listKey, singleKey)

	registerRows(func(v *Response[attrs]) ([]string, [][]string) {
		return []string{"ID"}, nil
	})

	expectPanic(t, "expected conflict panic when list handler is already registered", func() {
		registerRowsWithSingleResourceAdapter(func(v *Response[attrs]) ([]string, [][]string) {
			return []string{"ID"}, nil
		})
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistryRowsWithSingleResourceHelperNoPartialRegistrationWhenSingleDirectRegistered(t *testing.T) {
	type attrs struct {
		Name string `json:"name"`
	}

	listKey := reflect.TypeOf(&Response[attrs]{})
	singleKey := reflect.TypeOf(&SingleResponse[attrs]{})
	cleanupRegistryTypes(t, listKey, singleKey)

	registerDirect(func(v *SingleResponse[attrs], render func([]string, [][]string)) error {
		return nil
	})

	expectPanic(t, "expected conflict panic when single direct handler is already registered", func() {
		registerRowsWithSingleResourceAdapter(func(v *Response[attrs]) ([]string, [][]string) {
			return []string{"ID"}, nil
		})
	})

	assertRegistryTypeAbsent(t, listKey)
}

func TestOutputRegistryRowsWithSingleResourceHelperNoPartialRegistrationWhenListDirectRegistered(t *testing.T) {
	type attrs struct {
		Name string `json:"name"`
	}

	listKey := reflect.TypeOf(&Response[attrs]{})
	singleKey := reflect.TypeOf(&SingleResponse[attrs]{})
	cleanupRegistryTypes(t, listKey, singleKey)

	registerDirect(func(v *Response[attrs], render func([]string, [][]string)) error {
		return nil
	})

	expectPanic(t, "expected conflict panic when list direct handler is already registered", func() {
		registerRowsWithSingleResourceAdapter(func(v *Response[attrs]) ([]string, [][]string) {
			return []string{"ID"}, nil
		})
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistryRowsWithSingleResourceHelperNoPartialRegistrationWhenRowsNil(t *testing.T) {
	type attrs struct {
		Name string `json:"name"`
	}

	listKey := reflect.TypeOf(&Response[attrs]{})
	singleKey := reflect.TypeOf(&SingleResponse[attrs]{})
	cleanupRegistryTypes(t, listKey, singleKey)

	expectPanicContains(t, "nil rows function", func() {
		registerRowsWithSingleResourceAdapter[attrs](nil)
	})

	assertRegistryTypesAbsent(t, listKey, singleKey)
}

func TestOutputRegistryRowsWithSingleResourceHelperNilRowsPanicsBeforeConflictChecks(t *testing.T) {
	type attrs struct {
		Name string `json:"name"`
	}

	preregisterRowsForConflict[Response[attrs]](t, "ID")
	singleKey := reflect.TypeOf(&SingleResponse[attrs]{})
	cleanupRegistryTypes(t, singleKey)

	expectPanicContains(t, "nil rows function", func() {
		registerRowsWithSingleResourceAdapter[attrs](nil)
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistrySingleToListHelperRegistration(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
		if len(v.Data) == 0 {
			return []string{"value"}, nil
		}
		return []string{"value"}, [][]string{{v.Data[0]}}
	})

	key := reflect.TypeOf(&single{})
	cleanupRegistryTypes(t, key)

	handler := requireOutputHandler(t, key, "single-to-list helper")

	headers, rows, err := handler(&single{Data: "converted"})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if len(headers) != 1 || headers[0] != "value" {
		t.Fatalf("unexpected headers: %v", headers)
	}
	if len(rows) != 1 || len(rows[0]) != 1 || rows[0][0] != "converted" {
		t.Fatalf("unexpected rows: %v", rows)
	}
}

func TestOutputRegistryRowsWithSingleToListHelperRegistration(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	registerRowsWithSingleToListAdapter[single, list](func(v *list) ([]string, [][]string) {
		if len(v.Data) == 0 {
			return []string{"value"}, nil
		}
		return []string{"value"}, [][]string{{v.Data[0]}}
	})

	singleKey := reflect.TypeOf(&single{})
	listKey := reflect.TypeOf(&list{})
	cleanupRegistryTypes(t, singleKey, listKey)

	singleHandler := requireOutputHandler(t, singleKey, "single handler from rows+single-to-list helper")
	listHandler := requireOutputHandler(t, listKey, "list handler from rows+single-to-list helper")

	_, singleRows, err := singleHandler(&single{Data: "single-value"})
	if err != nil {
		t.Fatalf("single handler returned error: %v", err)
	}
	if len(singleRows) != 1 || len(singleRows[0]) != 1 || singleRows[0][0] != "single-value" {
		t.Fatalf("unexpected single rows: %v", singleRows)
	}

	_, listRows, err := listHandler(&list{Data: []string{"list-value"}})
	if err != nil {
		t.Fatalf("list handler returned error: %v", err)
	}
	if len(listRows) != 1 || len(listRows[0]) != 1 || listRows[0][0] != "list-value" {
		t.Fatalf("unexpected list rows: %v", listRows)
	}
}

func TestOutputRegistryRowsWithSingleToListHelperNoPartialRegistrationOnPanic(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	listKey := reflect.TypeOf(&list{})
	cleanupRegistryTypes(t, singleKey, listKey)

	registerRows(func(v *single) ([]string, [][]string) {
		return []string{"value"}, [][]string{{v.Data}}
	})

	expectPanic(t, "expected conflict panic when single handler is already registered", func() {
		registerRowsWithSingleToListAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})

	assertRegistryTypeAbsent(t, listKey)
}

func TestOutputRegistryRowsWithSingleToListHelperNoPartialRegistrationWhenListRegistered(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	listKey := reflect.TypeOf(&list{})
	cleanupRegistryTypes(t, singleKey, listKey)

	registerRows(func(v *list) ([]string, [][]string) {
		return []string{"value"}, nil
	})

	expectPanic(t, "expected conflict panic when list handler is already registered", func() {
		registerRowsWithSingleToListAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistryRowsWithSingleToListHelperNoPartialRegistrationWhenSingleDirectRegistered(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	listKey := reflect.TypeOf(&list{})
	cleanupRegistryTypes(t, singleKey, listKey)

	registerDirect(func(v *single, render func([]string, [][]string)) error {
		return nil
	})

	expectPanic(t, "expected conflict panic when single direct handler is already registered", func() {
		registerRowsWithSingleToListAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})

	assertRegistryTypeAbsent(t, listKey)
}

func TestOutputRegistryRowsWithSingleToListHelperNoPartialRegistrationWhenListDirectRegistered(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	listKey := reflect.TypeOf(&list{})
	cleanupRegistryTypes(t, singleKey, listKey)

	registerDirect(func(v *list, render func([]string, [][]string)) error {
		return nil
	})

	expectPanic(t, "expected conflict panic when list direct handler is already registered", func() {
		registerRowsWithSingleToListAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistryRowsWithSingleToListHelperNoPartialRegistrationWhenAdapterPanics(t *testing.T) {
	type single struct {
		Value string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	listKey := reflect.TypeOf(&list{})
	cleanupRegistryTypes(t, singleKey, listKey)

	expectPanic(t, "expected adapter panic for missing Data field", func() {
		registerRowsWithSingleToListAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})

	assertRegistryTypesAbsent(t, singleKey, listKey)
}

func TestOutputRegistrySingleToListHelperCopiesLinks(t *testing.T) {
	type single struct {
		Data  ResourceData
		Links Links
	}
	type list struct {
		Data  []ResourceData
		Links Links
	}

	registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
		if len(v.Data) == 0 {
			return []string{"id", "self"}, nil
		}
		return []string{"id", "self"}, [][]string{{v.Data[0].ID, v.Links.Self}}
	})

	key := reflect.TypeOf(&single{})
	cleanupRegistryTypes(t, key)

	handler := requireOutputHandler(t, key, "single-to-list links helper")

	headers, rows, err := handler(&single{
		Data: ResourceData{ID: "item-1", Type: ResourceType("items")},
		Links: Links{
			Self: "https://example.test/items/1",
		},
	})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if len(headers) != 2 || headers[0] != "id" || headers[1] != "self" {
		t.Fatalf("unexpected headers: %v", headers)
	}
	if len(rows) != 1 || len(rows[0]) != 2 {
		t.Fatalf("unexpected rows shape: %v", rows)
	}
	if rows[0][0] != "item-1" || rows[0][1] != "https://example.test/items/1" {
		t.Fatalf("unexpected row: %v", rows[0])
	}
}

func TestOutputRegistrySingleToListHelperPanicsWithoutDataField(t *testing.T) {
	type single struct {
		Value string
	}
	type list struct {
		Data []string
	}

	expectPanic(t, "expected panic when source Data field is missing", func() {
		registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, [][]string{{v.Data[0]}}
		})
	})
}

func TestOutputRegistrySingleToListHelperPanicsWhenSourceIsNotStruct(t *testing.T) {
	type single string
	type list struct {
		Data []string
	}

	expectPanicContains(t, "source type must be a struct", func() {
		registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
			return nil, nil
		})
	})
}

func TestOutputRegistrySingleToListHelperPanicsWhenTargetIsNotStruct(t *testing.T) {
	type single struct {
		Data string
	}
	type list []string

	expectPanicContains(t, "target type must be a struct", func() {
		registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
			return nil, nil
		})
	})
}

func TestOutputRegistrySingleToListHelperPanicsWhenTargetDataIsNotSlice(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data string
	}

	expectPanic(t, "expected panic when target Data field is not slice", func() {
		registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, [][]string{{v.Data}}
		})
	})
}

func TestOutputRegistrySingleToListHelperPanicsOnDataTypeMismatch(t *testing.T) {
	type single struct {
		Data int
	}
	type list struct {
		Data []string
	}

	expectPanic(t, "expected panic when Data element types mismatch", func() {
		registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})
}

func TestOutputRegistrySingleToListHelperPanicsOnNilRowsFunction(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	cleanupRegistryTypes(t, singleKey)

	expectPanicContains(t, "nil rows function", func() {
		registerSingleToListRowsAdapter[single, list](nil)
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistrySingleToListHelperAdapterValidationPanicsBeforeConflictChecks(t *testing.T) {
	type single struct {
		Value string
	}
	type list struct {
		Data []string
	}

	preregisterRowsForConflict[single](t, "value")

	expectPanicContains(t, "requires Data field", func() {
		registerSingleToListRowsAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})
}

func TestOutputRegistrySingleToListHelperNilRowsPanicsBeforeConflictChecks(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	preregisterRowsForConflict[single](t, "value")

	expectPanicContains(t, "nil rows function", func() {
		registerSingleToListRowsAdapter[single, list](nil)
	})
}

func TestOutputRegistryRegisterRowsPanicsOnDuplicate(t *testing.T) {
	type duplicate struct{}
	key := reflect.TypeOf(&duplicate{})
	cleanupRegistryTypes(t, key)

	registerRows(func(v *duplicate) ([]string, [][]string) {
		return []string{"value"}, [][]string{{"first"}}
	})

	expectPanic(t, "expected duplicate registration panic", func() {
		registerRows(func(v *duplicate) ([]string, [][]string) {
			return []string{"value"}, [][]string{{"second"}}
		})
	})
}

func TestOutputRegistryRegisterRowsPanicsOnNilFunction(t *testing.T) {
	type nilRows struct{}
	key := reflect.TypeOf(&nilRows{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil rows function", func() {
		registerRows[nilRows](nil)
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryRegisterRowsNilFunctionPanicsBeforeConflictChecks(t *testing.T) {
	type nilRows struct{}
	preregisterRowsForConflict[nilRows](t, "value")

	expectPanicContains(t, "nil rows function", func() {
		registerRows[nilRows](nil)
	})
}

func TestOutputRegistryRegisterRowsErrPanicsWhenDirectRegistered(t *testing.T) {
	type conflict struct{}
	key := reflect.TypeOf(&conflict{})
	cleanupRegistryTypes(t, key)

	registerDirect(func(v *conflict, render func([]string, [][]string)) error {
		return nil
	})

	expectPanic(t, "expected conflict panic when rowsErr registers after direct", func() {
		registerRowsErr(func(v *conflict) ([]string, [][]string, error) {
			return nil, nil, nil
		})
	})
}

func TestOutputRegistryRegisterRowsErrPanicsOnNilFunction(t *testing.T) {
	type nilRowsErr struct{}
	key := reflect.TypeOf(&nilRowsErr{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil rows function", func() {
		registerRowsErr[nilRowsErr](nil)
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryRegisterRowsErrNilFunctionPanicsBeforeConflictChecks(t *testing.T) {
	type nilRowsErr struct{}
	key := reflect.TypeOf(&nilRowsErr{})
	cleanupRegistryTypes(t, key)

	registerRowsErr(func(v *nilRowsErr) ([]string, [][]string, error) {
		return nil, nil, nil
	})

	expectPanicContains(t, "nil rows function", func() {
		registerRowsErr[nilRowsErr](nil)
	})
}

func TestOutputRegistryRegisterDirectPanicsWhenRowsRegistered(t *testing.T) {
	type conflict struct{}
	key := reflect.TypeOf(&conflict{})
	cleanupRegistryTypes(t, key)

	registerRows(func(v *conflict) ([]string, [][]string) {
		return []string{"value"}, [][]string{{"rows"}}
	})

	expectPanic(t, "expected conflict panic when direct registers after rows", func() {
		registerDirect(func(v *conflict, render func([]string, [][]string)) error {
			return nil
		})
	})
}

func TestOutputRegistryRegisterDirectPanicsOnNilFunction(t *testing.T) {
	type nilDirect struct{}
	key := reflect.TypeOf(&nilDirect{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "nil direct render function", func() {
		registerDirect[nilDirect](nil)
	})

	assertRegistryTypeAbsent(t, key)
}

func TestOutputRegistryRegisterDirectNilFunctionPanicsBeforeConflictChecks(t *testing.T) {
	type nilDirect struct{}
	key := reflect.TypeOf(&nilDirect{})
	cleanupRegistryTypes(t, key)

	registerDirect(func(v *nilDirect, render func([]string, [][]string)) error {
		return nil
	})

	expectPanicContains(t, "nil direct render function", func() {
		registerDirect[nilDirect](nil)
	})
}

func TestOutputRegistryRowsWithSingleToListHelperNoPartialRegistrationWhenRowsNil(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	listKey := reflect.TypeOf(&list{})
	cleanupRegistryTypes(t, singleKey, listKey)

	expectPanicContains(t, "nil rows function", func() {
		registerRowsWithSingleToListAdapter[single, list](nil)
	})

	assertRegistryTypesAbsent(t, singleKey, listKey)
}

func TestOutputRegistryRowsWithSingleToListHelperNilRowsPanicsBeforeConflictChecks(t *testing.T) {
	type single struct {
		Data string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	preregisterRowsForConflict[list](t, "value")
	cleanupRegistryTypes(t, singleKey)

	expectPanicContains(t, "nil rows function", func() {
		registerRowsWithSingleToListAdapter[single, list](nil)
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestOutputRegistryRowsWithSingleToListHelperAdapterValidationPanicsBeforeConflictChecks(t *testing.T) {
	type single struct {
		Value string
	}
	type list struct {
		Data []string
	}

	singleKey := reflect.TypeOf(&single{})
	preregisterRowsForConflict[list](t, "value")
	cleanupRegistryTypes(t, singleKey)

	expectPanicContains(t, "requires Data field", func() {
		registerRowsWithSingleToListAdapter[single, list](func(v *list) ([]string, [][]string) {
			return []string{"value"}, nil
		})
	})

	assertRegistryTypeAbsent(t, singleKey)
}

func TestEnsureRegistryTypesAvailablePanicsOnDuplicateTypes(t *testing.T) {
	type duplicate struct{}
	key := reflect.TypeOf(&duplicate{})
	cleanupRegistryTypes(t, key)

	expectPanicContains(t, "duplicate registration", func() {
		ensureRegistryTypesAvailable(key, key)
	})

	assertRegistryTypeAbsent(t, key)
}

func expectPanic(t *testing.T, message string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal(message)
		}
	}()
	fn()
}

func expectPanicContains(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		got := fmt.Sprint(r)
		if !strings.Contains(got, want) {
			t.Fatalf("panic %q does not contain %q", r, want)
		}
	}()
	fn()
}

func assertRowContains(t *testing.T, headers []string, rows [][]string, minColumns int, expected ...string) {
	t.Helper()
	if len(headers) == 0 || len(rows) == 0 {
		t.Fatalf("expected non-empty headers/rows, got headers=%v rows=%v", headers, rows)
	}
	if len(rows[0]) < minColumns {
		t.Fatalf("expected at least %d columns in row, got row=%v", minColumns, rows[0])
	}
	joined := strings.Join(rows[0], " ")
	for _, want := range expected {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected row to contain %q, got row=%v", want, rows[0])
		}
	}
}

func cleanupRegistryTypes(t *testing.T, types ...reflect.Type) {
	t.Helper()
	t.Cleanup(func() {
		for _, typ := range types {
			delete(outputRegistry, typ)
			delete(directRenderRegistry, typ)
		}
	})
}

func assertRegistryTypeAbsent(t *testing.T, typ reflect.Type) {
	t.Helper()
	if _, exists := outputRegistry[typ]; exists {
		t.Fatalf("registry type %v should be absent from output registry", typ)
	}
	if _, exists := directRenderRegistry[typ]; exists {
		t.Fatalf("registry type %v should be absent from direct render registry", typ)
	}
}

func assertRegistryTypesAbsent(t *testing.T, types ...reflect.Type) {
	t.Helper()
	for _, typ := range types {
		assertRegistryTypeAbsent(t, typ)
	}
}

func requireOutputHandler(t *testing.T, typ reflect.Type, label string) rowsFunc {
	t.Helper()
	handler, ok := outputRegistry[typ]
	if !ok || handler == nil {
		t.Fatalf("expected %s handler for type %v", label, typ)
	}
	return handler
}

func registerRowsForConflict[T any](headers ...string) {
	if len(headers) == 0 {
		headers = []string{"value"}
	}

	registerRows(func(*T) ([]string, [][]string) {
		return headers, nil
	})
}

func preregisterRowsForConflict[T any](t *testing.T, headers ...string) reflect.Type {
	t.Helper()

	key := reflect.TypeFor[*T]()
	cleanupRegistryTypes(t, key)
	registerRowsForConflict[T](headers...)
	return key
}
