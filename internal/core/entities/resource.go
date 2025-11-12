package entities

type Resource struct {
    ID            string                 `bson:"_id" json:"id"`
    Name          string                 `bson:"name" json:"name"`
    BlueprintName string                 `bson:"blueprint_name" json:"blueprintName"`
    Description   string                 `bson:"description" json:"description"`
    Status        ResourceStatus        `bson:"status" json:"status"`
    Spec          map[string]interface{} `bson:"spec" json:"spec"`
    Metadata      map[string]interface{} `bson:"metadata" json:"metadata"`
    CreatedAt     int64                  `bson:"created_at" json:"createdAt"`
    UpdatedAt     int64                  `bson:"updated_at" json:"updatedAt"`
}

type ResourceStatus struct {
    Conditions []Condition `json:"conditions"`
}
type Condition struct {
    LastTransitionTime   string `json:"lastTransitionTime"`
    ObservedGeneration   int    `json:"observedGeneration"`
    Reason               string `json:"reason"`
    Status               string `json:"status"`
    Type                 string `json:"type"`
}
