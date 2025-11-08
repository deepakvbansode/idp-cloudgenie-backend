package config

type MongoConfig struct {
	URI        string `required:"true" split_words:"true"`
	Database   string `required:"true" split_words:"true"`
	Collection string `required:"true" split_words:"true"`
}