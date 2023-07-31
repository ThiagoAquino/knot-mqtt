package utils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/CESARBR/knot-mqtt/internal/entities"
	"gopkg.in/yaml.v2"
)

type config interface {
	entities.Database | entities.Application | entities.Query | map[string]entities.Device | entities.IntegrationKNoTConfig | map[int]string
}

func readTextFile(filepathName string) ([]byte, error) {
	fileContent, err := ioutil.ReadFile(filepath.Clean(filepathName))
	return fileContent, err
}

func ConfigurationParser[T config](filepathName string, configEntity T) (T, error) {
	fileContent, err := readTextFile(filepath.Clean(filepathName))
	if err != nil {
		return configEntity, err
	}

	err = yaml.Unmarshal(fileContent, &configEntity)
	return configEntity, err
}
