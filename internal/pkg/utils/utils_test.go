package utils_test

import (
	"fmt"
	"testing"

	"github.com/huboh/gwatch/internal/pkg/utils"
)

func TestFind(t *testing.T) {
	type TestData[T comparable] struct {
		name      string
		data      []T
		result    T
		predicate func(e T, i int, s []T) bool
	}

	testData := []TestData[string]{
		{
			name:   "alpha",
			data:   []string{"one", "nine", "alpha"},
			result: "alpha",
			predicate: func(e string, i int, s []string) bool {
				return e == "alpha"
			},
		},
		{
			name:   "josephine",
			data:   []string{"john", "josephine", "andrew"},
			result: "josephine",
			predicate: func(e string, i int, s []string) bool {
				return e == "josephine"
			},
		},
	}

	for _, td := range testData {
		t.Run(fmt.Sprintf("Array Find \"%s\"", td.name), func(t *testing.T) {
			if result := utils.Find(td.data, td.predicate); td.result != result {
				t.Errorf("expected %s got %s\n", td.result, result)
			}
		})
	}
}
