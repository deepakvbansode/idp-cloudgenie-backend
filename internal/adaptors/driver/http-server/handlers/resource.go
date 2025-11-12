package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/constants"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/errors"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/usecases"
)

// extractIDFromPath extracts the last segment from the URL path as the resource ID
func extractIDFromPath(r *http.Request) string {
       parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
       if len(parts) > 0 {
               return parts[len(parts)-1]
       }
       return ""
}

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
	       ctx := r.Context()
	       id := extractIDFromPath(r)
	       if id == "" {
		       http.Error(w, "Missing resource id", http.StatusBadRequest)
		       return
	       }
	       resource, err := resourceService.GetResource(ctx, id)
	       if err != nil {
		       http.Error(w, err.Error(), http.StatusInternalServerError)
		       return
	       }
	       if resource == nil {
		       http.Error(w, "Resource not found", http.StatusNotFound)
		       return
	       }
	       w.Header().Set("Content-Type", "application/json")
	       w.WriteHeader(http.StatusOK)
	       json.NewEncoder(w).Encode(resource)
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
	       ctx := r.Context()
	       id := extractIDFromPath(r)
	       if id == "" {
		       http.Error(w, "Missing resource id", http.StatusBadRequest)
		       return
	       }
	       err := resourceService.DeleteResource(ctx, id)
	       if err != nil {
		       http.Error(w, err.Error(), http.StatusInternalServerError)
		       return
	       }
	       w.WriteHeader(http.StatusNoContent)
       }
}


func UpdateResourceStatusHandler(logger ports.Logger, resourceService *usecases.ResourceService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
	       ctx := r.Context()
	       id := extractIDFromPath(r)
	       if id == "" {
		       http.Error(w, "Missing resource id", http.StatusBadRequest)
		       return
	       }
	       var status entities.ResourceStatus
	       if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		       http.Error(w, "Invalid status body", http.StatusBadRequest)
		       return
	       }
	       err := resourceService.UpdateResourceStatus(ctx, id, status)
	       if err != nil {
		       http.Error(w, err.Error(), http.StatusInternalServerError)
		       return
	       }
	       w.WriteHeader(http.StatusOK)
       }
}
