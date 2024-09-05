package datastructures

import (
	"testing"
	"reflect"
)

func TestStringToInterfaceSlice(t *testing.T) {

	original := []string{"Test1", "Test2"}
	converted := StringToInterfaceSlice(original)

	if !reflect.DeepEqual(converted, []interface{}{"Test1", "Test2"}) {
		t.Fatalf("converted slice is not of type []interface{}")
	}
}