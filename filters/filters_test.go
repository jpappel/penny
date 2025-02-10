package filters_test

import (
	"github.com/jpappel/penny/filters"
	"testing"
)

type FiltererTestCase struct {
	name        string
	filterer    filters.Filterer
	input       []byte
	expected    []byte
	expectedErr error
}

func (tc FiltererTestCase) Test(t *testing.T) {
	result, err := tc.filterer.Filter(tc.input)
	if err != tc.expectedErr {
		t.Fatalf("Unexpected error: wanted `%v` got `%v`\n", tc.expectedErr, err)
	}

	rL := len(result)
	eL := len(tc.expected)
	if rL != eL {
		t.Errorf("Different buffer lengths: wanted %d got %d\n", eL, rL)
	}

	for i := range min(rL, eL) {
		if result[i] != tc.expected[i] {
			t.Error("Difference at index", i)
		}
	}

	if t.Failed() {
		t.Logf("Expected:\n======\n")
		t.Log(tc.expected)
		t.Logf("\n======\n\nResult\n======\n")
		t.Log(result)
	}
}

func TestFilterers(t *testing.T) {
	cases := []FiltererTestCase{
		{"Empty", filters.WordFilter{}, []byte("meep moop this is a test"), []byte("meep moop this is a test"), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.Test)
	}
}
