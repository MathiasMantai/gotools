package algorithm

import (
	"reflect"
	"testing"
)

var MergeSortTests = []struct {
	in  []int
	out []int
}{
	{[]int{3, 1, 2}, []int{1, 2, 3}},
	{[]int{70, 1, 150, 8}, []int{1, 8, 70, 150}},
	{[]int{5, 4, 3, 2, 1}, []int{1, 2, 3, 4, 5}},
}

func TestMergeSort(t *testing.T) {
	for _, test := range MergeSortTests {
		t.Run("test-merge-sort", func(t *testing.T) {
			actual := MergeSort(test.in)
			if !reflect.DeepEqual(actual, test.out) {
				t.Fatalf("Expected %v, got %v", test.out, actual)
			}
		})
	}
}
