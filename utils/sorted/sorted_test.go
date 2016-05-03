package sorted

import (
	"fmt"
	"sort"
	"testing"
)

func TestSortStringByLen(t *testing.T) {
	test := []string{"ccccc", "aa", "bbb"}
	want := []string{"aa", "bbb", "ccccc"}
	sort.Sort(ByStringLen(test))

	for i, v := range test {
		if v != want[i] {
			t.Errorf("SortLen Test Failed %+v", test)
		}
	}
}

func TestSortCharCount(t *testing.T) {
	test := []string{"/a/a/a/", "/c/c/c/c/c/c/c/c", "/b"}
	want := []string{"/c/c/c/c/c/c/c/c", "/a/a/a/", "/b"}
	bcc := ByCharCount(test, "/")
	sort.Sort(bcc)
	fmt.Printf("==%+v\n", test)

	q := sort.Reverse(sort.StringSlice(bcc.List))
	fmt.Println(q)

	for i, v := range bcc.List {
		if v != want[2-i] {
			t.Errorf("SortLen Test Failed %+v", test)
		}
	}
}

func TestSortByMapValue(t *testing.T) {

	orgin := map[string]float64{
		"a": 10.8,
		"b": 2.7,
		"c": 2.5,
	}

	test := make(map[string]interface{})

	for k, v := range orgin {
		test[k] = v
	}

	want := []string{"c", "b", "a"}
	bmv := ByMapValue(test)
	sort.Sort(bmv)

	for i, v := range bmv.List {
		if v != want[i] {
			t.Errorf("SortByMapValue Test Failed %+v != %+v", bmv.List, want)
		}
	}
}

func TestSortByMapValueLen(t *testing.T) {
	orgin := map[string][]int{
		"a": []int{1, 2, 3},
		"b": []int{3, 4},
		"c": []int{1, 2, 3, 4, 5},
	}

	test := make(map[string]interface{})
	for k, v := range orgin {
		test[k] = v
	}

	want := []string{"b", "a", "c"}

	bmvl := ByMapValueLen(test)
	sort.Sort(bmvl)

	for i, v := range bmvl.List {
		if v != want[i] {
			t.Errorf("SortByMapValueLen Test Slice Failed %+v != %+v", bmvl.List, want)
		}
	}

	mapOrgin := map[string]map[string]string{
		"a": map[string]string{"a": "b", "b": "a"},
		"b": map[string]string{"a": "b", "b": "a", "c": "a"},
		"c": map[string]string{"a": "b"},
	}

	mapTest := make(map[string]interface{})
	for k, v := range mapOrgin {
		mapTest[k] = v
	}

	want = []string{"c", "a", "b"}

	bmvl = ByMapValueLen(mapTest)

	sort.Sort(bmvl)
	for i, v := range bmvl.List {
		if v != want[i] {
			t.Errorf("SortByMapValueLen Test Map Failed %+v != %+v", bmvl.List, want)
		}
	}
}
