package config

type Config struct {
	// TODO: DELETE THIS, WE HAVE ENVS NOW
	GitHubApiTokenPath string `mapstructure:"github_api_token_path"`
}
