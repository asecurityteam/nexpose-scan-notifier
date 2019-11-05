package v1

import (
	"context"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
)

// DependencyCheckHandler stuff
type DependencyCheckHandler struct {
	DependencyChecker domain.DependencyChecker
}

// Handle stuff
func (h *DependencyCheckHandler) Handle(ctx context.Context) error {

	return nil
}
