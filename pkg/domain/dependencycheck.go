package domain

import "context"

// DependencyChecker represents an interface for checking external dependencies
type DependencyChecker interface {
	CheckDependencies(context.Context) error
}
