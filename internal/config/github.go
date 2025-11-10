package config

type GithubConfig struct{
	Token string `required:"true" split_words:"true"`
	Owner string `required:"true" split_words:"true"`
	Branch string `required:"false" split_words:"true" default:"main"`
}