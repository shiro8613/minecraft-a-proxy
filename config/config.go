package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var internal_config Config

func Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(b, &internal_config); err != nil {
		return err
	}

	return nil
}

func GetConfig() Config {
	return internal_config
}