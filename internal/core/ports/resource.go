package ports

import (
	"context"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
)

// RepositoryPort defines the interface for DB operations for resources
type RepositoryPort interface {
	SaveResource(ctx context.Context, resource *entities.Resource) (*entities.Resource, error)
	DeleteResource(ctx context.Context, id string) error
	GetResource(ctx context.Context, id string) (*entities.Resource, error)
	ListResources(ctx context.Context) ([]entities.Resource, error)
	UpdateResourceStatus(ctx context.Context, id string, status string) error
}

// GithubPort defines the interface for Github operations
type GithubPort interface {
	PushXRDToRepo(ctx context.Context, xrd string, repo string, path string) error
	// ...other github methods as needed
}
