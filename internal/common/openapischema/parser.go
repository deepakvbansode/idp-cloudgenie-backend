package openapischema

import "github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"

// ParseParameters recursively parses OpenAPI v3 schema properties and required fields into a map of Parameter.
// Pass the schema node where the parameters are defined (e.g., the 'properties' object for a spec).
func ParseParameters(schema map[string]interface{}) map[string]entities.Parameter {
	params := map[string]entities.Parameter{}
	if schema == nil {
		return params
	}
	// Get required array at this level
	requiredSet := map[string]bool{}
	if reqArr, ok := schema["required"].([]interface{}); ok {
		for _, reqName := range reqArr {
			if reqStr, ok := reqName.(string); ok {
				requiredSet[reqStr] = true
			}
		}
	}
	// Get properties
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		for pname, pval := range props {
			if pmap, ok := pval.(map[string]interface{}); ok {
				param := entities.Parameter{}
				if t, ok := pmap["type"].(string); ok {
					param.Type = t
				}
				if desc, ok := pmap["description"].(string); ok {
					param.Description = desc
				}
				param.Required = requiredSet[pname]
				// If this property is itself an object, you could recurse or flatten as needed
				params[pname] = param
			}
		}
	}
	return params
}
