package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/constants"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/errors"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/usecases"
)

func GetResourcesHandler(logger ports.Logger, resourceService *usecases.ResourceService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
	       ctx := r.Context()
	       resources, err := resourceService.ListResources(ctx)
	       if err != nil {
		       w.WriteHeader(http.StatusInternalServerError)
		       return
	       }
	       w.Header().Set("Content-Type", "application/json")
	       w.WriteHeader(http.StatusOK)
	       json.NewEncoder(w).Encode(resources)
       }
}

func GetResourceHandler(logger ports.Logger, resourceService *usecases.ResourceService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
	       // TODO: extract id from URL path
		   logger.Panic("GetResourceHandler not implemented yet")
	       w.WriteHeader(http.StatusNotImplemented)
       }
}

func CreateResourceHandler(logger ports.Logger, resourceService *usecases.ResourceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.WithField("tradeId", ctx.Value(constants.TraceIDKey))
		var resource entities.Resource
		if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
			log.Error("Failed to decode request body: ", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		createdResource, err := resourceService.CreateResource(ctx, &resource)
		if err != nil {
			log.Error("Failed to create resource: ", err)
			status := http.StatusInternalServerError
			switch err {
			case errors.ErrUnauthorized:
				status = http.StatusUnauthorized
			case errors.ErrForbidden:
				status = http.StatusForbidden
			case errors.ErrBlueprintNotFound, errors.ErrBlueprintNameMismatch, errors.ErrMissingRequiredFields, errors.ErrInvalidRequest:
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(createdResource); err != nil {
			log.Error("Failed to encode response: ", err)
		}
	}
}

func DeleteResourceHandler(logger ports.Logger, resourceService *usecases.ResourceService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
	       // TODO: extract id from URL path and call DeleteResource
	       w.WriteHeader(http.StatusNotImplemented)
       }
}

func UpdateResourceStatusHandler(logger ports.Logger,resourceService *usecases.ResourceService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
	       // TODO: extract id and status from URL/path/body and call UpdateResourceStatus
	       w.WriteHeader(http.StatusNotImplemented)
       }
}
