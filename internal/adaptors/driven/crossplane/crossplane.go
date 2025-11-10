package crossplane

import (
	"context"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/constants"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/k8s"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/openapischema"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/config"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// keysOfMap returns the keys of a map[string]interface{} as a []string
func keysOfMap(m map[string]interface{}) []string {
       keys := make([]string, 0, len(m))
       for k := range m {
	       keys = append(keys, k)
       }
	return keys
}



// CrossplaneAdaptor implements the CrossplanePort interface
type CrossplaneAdaptor struct {
	logger ports.Logger
	config config.CrossplaneConfig
	// Add any necessary fields for Crossplane integration, e.g., API client
}

// ListXRDsWithLabel lists all XRDs (CompositeResourceDefinitions) with the given label selector and returns the raw JSON result
func (cp *CrossplaneAdaptor) listXRDsWithLabel(ctx context.Context, labelSelector string) ([]byte, error) {
       // Assumes running in-cluster or with KUBECONFIG set
       config, err := k8s.GetKubeConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get k8s config: %w", err)
		}

       dynClient, err := dynamic.NewForConfig(config)
       if err != nil {
	       return nil, fmt.Errorf("failed to create dynamic client: %w", err)
       }

       gvr := schema.GroupVersionResource{
	       Group:    "apiextensions.crossplane.io",
	       Version:  "v1",
	       Resource: "compositeresourcedefinitions",
       }

       list, err := dynClient.Resource(gvr).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
       if err != nil {
	       return nil, fmt.Errorf("failed to list XRDs: %w", err)
       }

       return json.Marshal(list)
}

// NewCrossplaneAdaptor creates a new instance of CrossplaneAdaptor
func NewCrossplaneAdaptor(logger ports.Logger, config config.CrossplaneConfig) *CrossplaneAdaptor {
	return &CrossplaneAdaptor{
		logger: logger,
		config: config,
	}
}

// ListBlueprints retrieves the list of blueprints from Crossplane
func (cp *CrossplaneAdaptor) ListBlueprints(ctx context.Context) ([]entities.Blueprint, error) {
       log := cp.logger.WithField("tradeId", ctx.Value(constants.TraceIDKey))
       log.Info("Listing blueprints from Crossplane")
       xrdsBytes, err := cp.listXRDsWithLabel(ctx, "blueprint-name")
       if err != nil {
	       log.Error("Failed to list XRDs: ", err)
	       return nil, err
       }
       log.Info("Fetched XRDs: ", string(xrdsBytes))

       // Parse the returned XRDs JSON (as unstructured)
       var xrdList struct {
	       Items []map[string]interface{} `json:"items"`
       }
       if err := json.Unmarshal(xrdsBytes, &xrdList); err != nil {
	       log.Error("Failed to unmarshal XRDs: ", err)
	       return nil, err
       }

       var blueprints []entities.Blueprint
       for _, item := range xrdList.Items {
	       metadata, _ := item["metadata"].(map[string]interface{})
	       spec, _ := item["spec"].(map[string]interface{})
	       kind := ""
	       name := ""
	       if metadata != nil {
		       if nameVal, ok := metadata["name"].(string); ok {
			       kind = nameVal
		       }
		       // name from label 'blueprint-name'
		       if labels, ok := metadata["labels"].(map[string]interface{}); ok {
			       if n, ok := labels["blueprint-name"].(string); ok {
				       name = n
			       }
		       }
	       }
	       description := ""
	       category := ""
	       version := ""
		parameters := map[string]entities.Parameter{}

	       // Extract category from group
	       if spec != nil {
		       if cat, ok := spec["group"].(string); ok {
			       category = cat
		       }
		       // Find version object with Referenceable=true
		       if verArr, ok := spec["versions"].([]interface{}); ok && len(verArr) > 0 {
			       var refVer map[string]interface{}
			       for _, v := range verArr {
				       if vmap, ok := v.(map[string]interface{}); ok {
					       if ref, ok := vmap["referenceable"].(bool); ok && ref {
						       refVer = vmap
						       break
					       }
				       }
			       }
			       if refVer == nil {
				       // fallback to first
				       if vmap, ok := verArr[0].(map[string]interface{}); ok {
					       refVer = vmap
				       }
			       }
			       if refVer != nil {
				       if vstr, ok := refVer["name"].(string); ok {
					       version = vstr
				       }
				       // Try to extract description and parameters from schema
		       if schema, ok := refVer["schema"].(map[string]interface{}); ok {
			       openAPIV3Schema, _ := schema["openAPIV3Schema"].(map[string]interface{})
			       if openAPIV3Schema != nil {
				       // description from openAPIV3Schema.description
				       if desc, ok := openAPIV3Schema["description"].(string); ok {
					       description = desc
				       }
				       // parameters: look for openAPIV3Schema.properties.spec.properties (treat all as parameters)
				       if props, ok := openAPIV3Schema["properties"].(map[string]interface{}); ok {
					       if specProp, ok := props["spec"].(map[string]interface{}); ok {
						       if specProps, ok := specProp["properties"].(map[string]interface{}); ok {
							       // DEBUG: Log available keys at this level
							       cp.logger.Info("spec.properties keys: ", keysOfMap(specProps))
															   // Treat all fields under spec.properties as parameters using the shared parser
															   parameters = openapischema.ParseParameters(specProp)
						       } else {
							       cp.logger.Warn("No 'properties' found under spec. Available keys: ", keysOfMap(specProp))
						       }
					       } else {
						       cp.logger.Warn("No 'spec' property found in openAPIV3Schema.properties. Available keys: ", keysOfMap(props))
					       }
				       } else {
					       cp.logger.Warn("No 'properties' found in openAPIV3Schema. Available keys: ", keysOfMap(openAPIV3Schema))
				       }
			       }
		       }

			       }
		       }
	       }

	       blueprints = append(blueprints, entities.Blueprint{
		       Kind:          kind,
		       Name:        name,
		       Description: description,
		       Parameters:  parameters,
		       Category:    category,
		       Version:     version,
	       })
       }
       return blueprints, nil
}

// BuildXRD builds an XRD YAML from a resource and blueprint, validating required fields
func (cp *CrossplaneAdaptor) BuildXRD(ctx context.Context,resource *entities.Resource, blueprint *entities.Blueprint) (string, error) {
	requiredMissing := []string{}
	filteredSpec := map[string]interface{}{}
	for pname, param := range blueprint.Parameters {
		val, ok := resource.Spec[pname]
		if param.Required && (!ok || val == nil || (param.Type == "string" && val == "")) {
			requiredMissing = append(requiredMissing, pname)
		}
		if ok {
			filteredSpec[pname] = val
		}
	}
	if len(requiredMissing) > 0 {
		return "", fmt.Errorf("missing required fields in spec: %v", requiredMissing)
	}
	xrd := map[string]interface{}{
		"apiVersion": fmt.Sprintf("%s/%s", blueprint.Category, blueprint.Version),
		"kind": blueprint.Kind,
		"metadata": map[string]interface{}{
			"name": resource.Name,
			"annotations": map[string]interface{}{
				"description": resource.Description,
			},
		},
		"spec": filteredSpec,
	}
	xrdYAML, err := yaml.Marshal(xrd)
	if err != nil {
		return "", fmt.Errorf("failed to marshal XRD to YAML: %w", err)
	}
	return string(xrdYAML), nil
}