package usecases

import (
	"context"
	"fmt"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/errors"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
)

type ResourceService struct {
	logger ports.Logger
	githubProvider ports.GithubPort
	repository ports.RepositoryPort
	crossplane ports.CrossplanePort
}

func NewResourceService(logger ports.Logger, githubProvider ports.GithubPort, repository ports.RepositoryPort, crossplane ports.CrossplanePort) *ResourceService {
	return &ResourceService{
		logger:      logger,
		githubProvider: githubProvider,
		repository:  repository,
		crossplane:  crossplane,
	}
}
		

func (s *ResourceService) CreateResource(ctx context.Context, resource *entities.Resource) (*entities.Resource, error) {
	// 1. Validate the user (skipped for now)

	// 2. Fetch blueprints from Crossplane and validate blueprintName
	blueprints, err := s.crossplane.ListBlueprints(ctx)
	if err != nil {
		return nil, err
	}
	var blueprint *entities.Blueprint
	for _, bp := range blueprints {
		if bp.Name == resource.BlueprintName {
			blueprint = &bp
			break
		}
	}
	if blueprint == nil {
		return nil, errors.ErrBlueprintNotFound
	}

	// 3. Build XRD YAML using CrossplaneAdaptor (handles validation and spec filtering)
	xrdYAML, err := s.crossplane.BuildXRD(ctx, resource, blueprint)
	if err != nil {
		return nil, err
	}

	// 4. Save the metadata in db (repository)
	savedResource, err := s.repository.SaveResource(ctx, resource)
	if err != nil {
		return nil, err
	}

	// 5. Push the XRD to github repo
	repoName := "idp-cloudgenie-state"
	xrdPath := fmt.Sprintf("resources/%s/%s.yaml", resource.BlueprintName, resource.Name)
	err = s.githubProvider.PushXRDToRepo(ctx, xrdYAML, repoName, xrdPath)
	if err != nil {
		return nil, err
	}

	return savedResource, nil
}


func (s *ResourceService) UpdateResource(ctx context.Context,resource *entities.Resource) (*entities.Resource, error) {
	// Save updated resource in db (repository)
	updatedResource, err := s.repository.SaveResource(ctx, resource)
	if err != nil {
		 return nil, err
	}

	// Optionally update XRD in github if needed (placeholder logic)
	// xrd := "updated-xrd-content" // TODO: generate updated XRD if required
	// err = s.githubProvider.PushXRDToRepo(xrd, "repo-name", "path/to/xrd.yaml")
	// if err != nil {
	//      return nil, err
	// }

	return updatedResource, nil
}


func (s *ResourceService) DeleteResource(ctx context.Context,id string) error {
	// Delete resource from db (repository)
	err := s.repository.DeleteResource(ctx, id)
	if err != nil {
		 return err
	}

	// Optionally delete XRD from github (not implemented)
	// err = s.githubProvider.DeleteXRDFromRepo("repo-name", "path/to/xrd.yaml")
	// if err != nil {
	//      return err
	// }

	return nil
}


func (s *ResourceService) GetResource(ctx context.Context,id string) (*entities.Resource, error) {
	// Get resource from db (repository)
	return s.repository.GetResource(ctx, id)
}


func (s *ResourceService) ListResources(ctx context.Context) ([]entities.Resource, error) {
	// List all resources from db (repository)
	return s.repository.ListResources(ctx)
}


func (s *ResourceService) UpdateResourceStatus(ctx context.Context, resourceName string, status entities.ResourceStatus) error {
	// Update status in db (repository)
	return s.repository.UpdateResourceStatus(ctx, resourceName, status)
}
