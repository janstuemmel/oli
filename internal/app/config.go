package app

type ConfigRoles struct {
	Name   string `mapstructure:"name"`
	System string `mapstructure:"system"`
	Model  string `mapstructure:"model"`
}

type Config struct {
	Apikey string `mapstructure:"apikey"`
	Model  string `mapstructure:"model"`
	Online bool
	System string        `mapstructure:"system"`
	Pipe   string        `mapstructure:"pipe"`
	Roles  []ConfigRoles `mapstructure:"roles"`
}
