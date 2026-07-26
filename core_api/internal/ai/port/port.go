// Package port defines the narrow module dependencies required by AI services.
package port

import (
	"context"
	"io"

	"github.com/1024XEngineer/Holonic-Asset/internal/ai/domain"
)

type ProjectContextReader interface {
	GetProjectContext(ctx context.Context, projectID uint) (*domain.ProjectContext, error)
}

type MediaSource struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
}

type MediaImport struct {
	Body        io.Reader
	ContentType string
	Filename    string
}

// MediaGateway resolves stable media references and imports provider outputs.
type MediaGateway interface {
	Open(ctx context.Context, ref domain.MediaRef) (*MediaSource, error)
	Import(ctx context.Context, request *MediaImport) (*domain.MediaRef, error)
}

type UsageRepository interface {
	SaveInvocation(ctx context.Context, invocation *domain.Invocation) error
}
