package utils

import (
	"reflect"
	"testing"
)

func TestSliceFilter_Strings_RemovesMatching(t *testing.T) {
	src := []string{"a", "b", "c", "d"}
	filter := []string{"b", "d"}
	result := SliceFilter(src, filter)
	expected := []string{"a", "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSliceFilter_Ints_RemovesMatching(t *testing.T) {
	src := []int{1, 2, 3, 4, 5}
	filter := []int{2, 4}
	result := SliceFilter(src, filter)
	expected := []int{1, 3, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSliceFilter_EmptySource(t *testing.T) {
	src := []int{}
	filter := []int{1, 2}
	result := SliceFilter(src, filter)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestSliceFilter_EmptyFilter(t *testing.T) {
	src := []string{"x", "y"}
	filter := []string{}
	result := SliceFilter(src, filter)
	expected := []string{"x", "y"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSliceFilter_RemoveAll(t *testing.T) {
	src := []int{1, 2, 3}
	filter := []int{1, 2, 3}
	result := SliceFilter(src, filter)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}
