package mongo

import (
	"context"
	"time"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/config"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RepositoryAdaptor struct {
    logger     ports.Logger
    config     config.MongoConfig
    client     *mongo.Client
    collection *mongo.Collection
}
func NewRepositoryAdaptor(logger ports.Logger, config config.MongoConfig) *RepositoryAdaptor {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI))
	if err != nil {
		logger.Panic("Failed to connect to MongoDB:", err)
		return nil
	}
	collection := client.Database(config.Database).Collection(config.Collection)
	return &RepositoryAdaptor{
		logger:     logger,
		config:     config,
		client:     client,
		collection: collection,
	}
}

func (r *RepositoryAdaptor) SaveResource(ctx context.Context,resource *entities.Resource) (*entities.Resource, error) {
	resource.ID = resource.Name
	_, err := r.collection.InsertOne(ctx, resource)
	if err != nil {
		r.logger.Error("Failed to insert resource:", err)
		return nil, err
	}
	return resource, nil
}


func (r *RepositoryAdaptor) DeleteResource(ctx context.Context,id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete resource:", err)
		return err
	}
	return nil
}

func (r *RepositoryAdaptor) GetResource(ctx context.Context,id string) (*entities.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"_id": id}
	var resource entities.Resource
	err := r.collection.FindOne(ctx, filter).Decode(&resource)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error("Failed to get resource:", err)
		return nil, err
	}
	return &resource, nil
}

func (r *RepositoryAdaptor) ListResources(ctx context.Context,) ([]entities.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		r.logger.Error("Failed to list resources:", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	var resources []entities.Resource
	for cursor.Next(ctx) {
		var resource entities.Resource
		if err := cursor.Decode(&resource); err != nil {
			r.logger.Error("Failed to decode resource:", err)
			continue
		}
		resources = append(resources, resource)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error:", err)
		return nil, err
	}
	return resources, nil
}



func (r *RepositoryAdaptor) UpdateResourceStatus(ctx context.Context,id string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status, "updated_at": time.Now().Unix()}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update resource status:", err)
		return err
	}
	return nil
}