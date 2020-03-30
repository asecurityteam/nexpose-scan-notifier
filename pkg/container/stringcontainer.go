package container

import "strings"

// StringContainer stores a slice of strings for efficient membership checks.
type StringContainer map[string]struct{}

// Contains efficiently returns whether the given string exists in the container.
func (c StringContainer) Contains(item string) bool {
	if _, ok := c[item]; !ok {
		return false
	}
	return true
}

// String returns a flattened string reprentation of the keys in the StringContainer.
func (c StringContainer) String() string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return "[" + strings.Join(keys, ", ") + "]"
}

// NewStringContainer returns a StringContainer initialized with the given strings.
func NewStringContainer(items []string) *StringContainer {
	container := make(StringContainer)
	for _, item := range items {
		container[item] = struct{}{}
	}
	return &container
}
