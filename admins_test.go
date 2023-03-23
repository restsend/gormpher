package gormpher

import (
	"reflect"
	"testing"
)

func TestConvertTypeToJS(t *testing.T) {
	testCases := []struct {
		input reflect.Kind
		want  string
	}{
		{reflect.String, "string"},
		{reflect.Int, "number"},
		{reflect.Float64, "number"},
		{reflect.Bool, "boolean"},
		{reflect.Struct, "object"},
		{reflect.Map, "object"},
		{reflect.Slice, "object"},
		{reflect.Ptr, "any"},
		{reflect.Interface, "any"},
		{reflect.Complex128, "unknown"},
	}

	for _, tc := range testCases {
		got := convertGoTypeToJS(tc.input)
		if got != tc.want {
			t.Errorf("convertTypeToJS(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
