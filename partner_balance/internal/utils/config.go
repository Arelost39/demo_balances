package utils

import (
	"os"
	"gopkg.in/yaml.v3"
)

type PartnerConfig struct {
	Token       string `yaml:"token"`
	Description string `yaml:"description"`
	IsActive    bool   `yaml:"is_active"`
}

type Config struct {
	Networks map[string]map[string]PartnerConfig `yaml:"networks"`
}

var AppConfig Config

func LoadConfig(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &AppConfig)
	if err != nil {
		return err
	}

	return nil
}