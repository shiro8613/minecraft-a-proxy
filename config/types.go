package config

type Config struct {
	Bind string `yaml:"bind"`
	Servers map[string]string `yaml:"servers"`
}
