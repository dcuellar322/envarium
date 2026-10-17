package application

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTypedResolutionCopyIsolatesEveryReferenceField(t *testing.T) {
	t.Parallel()
	var original EffectiveManifest
	fillResolutionFixture(reflect.ValueOf(&original).Elem())
	before, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	copied := cloneResolution(original)
	if !reflect.DeepEqual(original, copied) {
		t.Fatal("copy changed values or nil/empty distinctions")
	}
	mutateResolutionFixture(reflect.ValueOf(&copied).Elem())
	after, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("caller mutation leaked into the retained manifest/provenance")
	}
	if reflect.DeepEqual(original, copied) {
		t.Fatal("fixture did not mutate")
	}
	if !reflect.DeepEqual(EffectiveManifest{}, cloneResolution(EffectiveManifest{})) {
		t.Fatal("nil references changed")
	}
}

// Reflection is confined to this test: new reference fields automatically join
// the isolation check instead of silently escaping a hand-picked fixture.
func fillResolutionFixture(v reflect.Value) {
	//nolint:exhaustive // Unsupported fixture kinds fail explicitly instead of being silently skipped.
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fillResolutionFixture(v.Field(i))
		}
	case reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
		fillResolutionFixture(v.Elem())
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		fillResolutionFixture(v.Index(0))
	case reflect.Map:
		v.Set(reflect.MakeMap(v.Type()))
		child := reflect.New(v.Type().Elem()).Elem()
		fillResolutionFixture(child)
		v.SetMapIndex(reflect.ValueOf("fixture"), child)
	case reflect.Interface:
		v.Set(reflect.ValueOf(map[string]any{"nested": []any{map[string]any{"key": "value"}}}))
	case reflect.String:
		v.SetString("fixture")
	case reflect.Int:
		v.SetInt(1)
	case reflect.Bool:
		v.SetBool(true)
	default:
		panic("unhandled manifest field in isolation fixture: " + v.Type().String())
	}
}

func mutateResolutionFixture(v reflect.Value) {
	//nolint:exhaustive // Unsupported fixture kinds fail explicitly instead of being silently skipped.
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			mutateResolutionFixture(v.Field(i))
		}
	case reflect.Pointer:
		if !v.IsNil() {
			mutateResolutionFixture(v.Elem())
		}
	case reflect.Interface:
		if !v.IsNil() {
			child := reflect.New(v.Elem().Type()).Elem()
			child.Set(v.Elem())
			mutateResolutionFixture(child)
			if v.CanSet() {
				v.Set(child)
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			mutateResolutionFixture(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			child := reflect.New(v.Type().Elem()).Elem()
			child.Set(v.MapIndex(key))
			mutateResolutionFixture(child)
			v.SetMapIndex(key, child)
		}
	case reflect.String:
		if v.CanSet() {
			v.SetString("changed")
		}
	case reflect.Int:
		if v.CanSet() {
			v.SetInt(2)
		}
	case reflect.Bool:
		if v.CanSet() {
			v.SetBool(false)
		}
	default:
		panic("unhandled manifest mutation field: " + v.Type().String())
	}
}
