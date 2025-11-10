package ports

import (
	"context"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
)

// CrossplanePort defines the interface for blueprint operations via Crossplane
type CrossplanePort interface {
	ListBlueprints(ctx context.Context) ([]entities.Blueprint, error)
	BuildXRD(ctx context.Context, resource *entities.Resource, blueprint *entities.Blueprint) (string, error)
}
