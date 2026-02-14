package asc

import (
	"fmt"
	"reflect"
)

// rowsFunc extracts headers and rows from a typed response value.
type rowsFunc func(data any) ([]string, [][]string, error)

// directRenderFunc renders the value using the provided render callback.
// Used for multi-table types that need to call render more than once.
type directRenderFunc func(data any, render func([]string, [][]string)) error

// outputRegistry maps concrete pointer types to their rows-extraction function.
var outputRegistry = map[reflect.Type]rowsFunc{}

// directRenderRegistry maps types that need direct render control (multi-table output).
var directRenderRegistry = map[reflect.Type]directRenderFunc{}

func ensureRegistryTypeAvailable(t reflect.Type) {
	if _, exists := outputRegistry[t]; exists {
		panic(fmt.Sprintf("output registry: duplicate registration for %s", t))
	}
	if _, exists := directRenderRegistry[t]; exists {
		panic(fmt.Sprintf("output registry: duplicate registration for %s", t))
	}
}

func ensureRegistryTypesAvailable(types ...reflect.Type) {
	seen := make(map[reflect.Type]struct{}, len(types))
	for _, t := range types {
		if _, exists := seen[t]; exists {
			panic(fmt.Sprintf("output registry: duplicate registration for %s", t))
		}
		seen[t] = struct{}{}
		ensureRegistryTypeAvailable(t)
	}
}

// registerRows registers a rows function for the given pointer type.
// The function must accept a pointer and return (headers, rows).
func registerRows[T any](fn func(*T) ([]string, [][]string)) {
	t := reflect.TypeFor[*T]()
	ensureRegistryTypeAvailable(t)
	outputRegistry[t] = func(data any) ([]string, [][]string, error) {
		h, r := fn(data.(*T))
		return h, r, nil
	}
}

// registerRowsErr registers a rows function that can return an error.
func registerRowsErr[T any](fn func(*T) ([]string, [][]string, error)) {
	t := reflect.TypeFor[*T]()
	ensureRegistryTypeAvailable(t)
	outputRegistry[t] = func(data any) ([]string, [][]string, error) {
		return fn(data.(*T))
	}
}

func registerSingleLinkageRows[T any](extract func(*T) ResourceData) {
	registerRows(func(v *T) ([]string, [][]string) {
		return linkagesRows(&LinkagesResponse{Data: []ResourceData{extract(v)}})
	})
}

func registerIDStateRows[T any](extract func(*T) (string, string), rows func(string, string) ([]string, [][]string)) {
	registerRows(func(v *T) ([]string, [][]string) {
		id, state := extract(v)
		return rows(id, state)
	})
}

func registerIDBoolRows[T any](extract func(*T) (string, bool), rows func(string, bool) ([]string, [][]string)) {
	registerRows(func(v *T) ([]string, [][]string) {
		id, deleted := extract(v)
		return rows(id, deleted)
	})
}

func registerResponseDataRows[T any](rows func([]Resource[T]) ([]string, [][]string)) {
	registerRows(func(v *Response[T]) ([]string, [][]string) {
		return rows(v.Data)
	})
}

// registerSingleResourceRowsAdapter registers rows rendering for list renderers
// by adapting SingleResponse[T] into Response[T] with one item in Data.
func registerSingleResourceRowsAdapter[T any](rows func(*Response[T]) ([]string, [][]string)) {
	registerRows(func(v *SingleResponse[T]) ([]string, [][]string) {
		return rows(&Response[T]{Data: []Resource[T]{v.Data}})
	})
}

// registerRowsWithSingleResourceAdapter registers both list and single handlers
// for row renderers that operate on Response[T].
func registerRowsWithSingleResourceAdapter[T any](rows func(*Response[T]) ([]string, [][]string)) {
	ensureRegistryTypesAvailable(
		reflect.TypeFor[*Response[T]](),
		reflect.TypeFor[*SingleResponse[T]](),
	)
	registerRows(rows)
	registerSingleResourceRowsAdapter(rows)
}

// registerSingleToListRowsAdapter registers rows rendering by adapting a single
// response struct into a corresponding list response struct using shared field
// names. The source type must expose `Data` and may expose `Links`; the target
// type must expose `Data` as a slice and may expose `Links`.
func registerSingleToListRowsAdapter[T any, U any](rows func(*U) ([]string, [][]string)) {
	registerRows(singleToListRowsAdapter[T, U](rows))
}

func singleToListRowsAdapter[T any, U any](rows func(*U) ([]string, [][]string)) func(*T) ([]string, [][]string) {
	sourceType := reflect.TypeFor[T]()
	targetType := reflect.TypeFor[U]()
	if sourceType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("output registry: single/list adapter source type must be a struct: %s", sourceType))
	}
	if targetType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("output registry: single/list adapter target type must be a struct: %s", targetType))
	}

	sourceDataField, sourceHasData := sourceType.FieldByName("Data")
	targetDataField, targetHasData := targetType.FieldByName("Data")
	if !sourceHasData || !targetHasData {
		panic("output registry: single/list adapter requires Data field on source and target")
	}
	if targetDataField.Type.Kind() != reflect.Slice {
		panic("output registry: single/list adapter target Data field must be a slice")
	}
	targetElemType := targetDataField.Type.Elem()
	if !sourceDataField.Type.AssignableTo(targetElemType) {
		panic(fmt.Sprintf(
			"output registry: adapter Data type mismatch source=%s target=%s",
			sourceDataField.Type,
			targetElemType,
		))
	}

	sourceLinksField, sourceHasLinks := sourceType.FieldByName("Links")
	targetLinksField, targetHasLinks := targetType.FieldByName("Links")
	copyLinks := sourceHasLinks &&
		targetHasLinks &&
		sourceLinksField.Type.AssignableTo(targetLinksField.Type)

	return func(v *T) ([]string, [][]string) {
		source := reflect.ValueOf(v).Elem()
		var target U
		targetValue := reflect.ValueOf(&target).Elem()

		sourceData := source.FieldByIndex(sourceDataField.Index)
		targetData := targetValue.FieldByIndex(targetDataField.Index)

		rowsSlice := reflect.MakeSlice(targetData.Type(), 1, 1)
		rowsSlice.Index(0).Set(sourceData)
		targetData.Set(rowsSlice)

		if copyLinks {
			sourceLinks := source.FieldByIndex(sourceLinksField.Index)
			targetLinks := targetValue.FieldByIndex(targetLinksField.Index)
			targetLinks.Set(sourceLinks)
		}

		return rows(&target)
	}
}

// registerRowsWithSingleToListAdapter registers both list and single handlers
// when list rendering expects a concrete list response type.
func registerRowsWithSingleToListAdapter[T any, U any](rows func(*U) ([]string, [][]string)) {
	ensureRegistryTypesAvailable(
		reflect.TypeFor[*U](),
		reflect.TypeFor[*T](),
	)
	adapter := singleToListRowsAdapter[T, U](rows)
	registerRows(rows)
	registerRows(adapter)
}

// registerDirect registers a type that needs direct render control (multi-table output).
func registerDirect[T any](fn func(*T, func([]string, [][]string)) error) {
	t := reflect.TypeFor[*T]()
	ensureRegistryTypeAvailable(t)
	directRenderRegistry[t] = func(data any, render func([]string, [][]string)) error {
		return fn(data.(*T), render)
	}
}

// renderByRegistry looks up the rows function for the given value and renders
// using the provided render function (RenderTable or RenderMarkdown).
// Falls back to JSON output for unregistered types.
func renderByRegistry(data any, render func([]string, [][]string)) error {
	t := reflect.TypeOf(data)

	// Check direct render registry first (multi-table types).
	if fn, ok := directRenderRegistry[t]; ok {
		return fn(data, render)
	}

	// Standard single-table types.
	if fn, ok := outputRegistry[t]; ok {
		h, r, err := fn(data)
		if err != nil {
			return err
		}
		render(h, r)
		return nil
	}

	return PrintJSON(data)
}
