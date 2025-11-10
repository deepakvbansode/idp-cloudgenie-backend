package entities

type Parameter struct{
	Description string `bson:"description" json:"description"`
	Type        string `bson:"type" json:"type"`
	Required    bool   `bson:"required" json:"required"`
}
type Blueprint struct {
	Kind       string            `bson:"kind" json:"kind"`
	Name        string            `bson:"name" json:"name"`
	Description string            `bson:"description" json:"description"`
	Parameters  map[string]Parameter `bson:"parameters" json:"parameters"`
	Category    string            `bson:"category" json:"category"` // frontend or backend
	Version     string            `bson:"version" json:"version"`
}
//crossplane blueprint entity