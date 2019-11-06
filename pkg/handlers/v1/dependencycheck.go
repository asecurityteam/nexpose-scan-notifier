package v1

import (
	"context"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
)

// DependencyCheckHandler takes in a domain.DependencyChecker to check external dependencies
type DependencyCheckHandler struct {
	DependencyChecker domain.DependencyChecker
}

// Handle makes a call CheckDependencies from DependencyChecker that verifies this
// app can talk to it's external dependencies
func (h *DependencyCheckHandler) Handle(ctx context.Context) error {
	return h.DependencyChecker.CheckDependencies(ctx)
}
