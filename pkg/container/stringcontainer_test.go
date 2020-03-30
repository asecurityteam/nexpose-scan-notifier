package container

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStringContainer(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected *StringContainer
	}{
		{
			name:  "slice of strings",
			items: []string{"foo", "bar"},
			expected: &StringContainer{
				"foo": struct{}{},
				"bar": struct{}{},
			},
		},
		{
			name:     "empty slice",
			items:    []string{},
			expected: &StringContainer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewStringContainer(tt.items)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name      string
		container *StringContainer
		item      string
		expected  bool
	}{
		{
			name: "true case",
			container: &StringContainer{
				"foo": struct{}{},
				"bar": struct{}{},
			},
			item:     "foo",
			expected: true,
		},
		{
			name: "false case",
			container: &StringContainer{
				"foo": struct{}{},
			},
			item:     "bar",
			expected: false,
		},
		{
			name:      "empty false case",
			container: &StringContainer{},
			item:      "foo",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.container.Contains(tt.item)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name           string
		container      *StringContainer
		expectedString string
		// string value is non-deterministic for StringContainers with len > 1
		expectedValues []string
	}{
		{
			name:           "empty",
			container:      &StringContainer{},
			expectedString: "[]",
		},
		{
			name: "one item",
			container: &StringContainer{
				"foo": struct{}{},
			},
			expectedString: "[foo]",
		},
		{
			name: "more than one item",
			container: &StringContainer{
				"what's":   struct{}{},
				"taters":   struct{}{},
				"precious": struct{}{},
			},
			expectedValues: []string{"what's", "taters", "precious"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := fmt.Sprintf("%s", tt.container)
			if tt.expectedValues != nil {
				for _, item := range tt.expectedValues {
					require.Contains(t, actual, item)
				}
			} else {
				require.Equal(t, tt.expectedString, actual)
			}
		})
	}
}
