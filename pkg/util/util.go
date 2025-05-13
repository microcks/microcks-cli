package util

import (
	"os"

	"gopkg.in/yaml.v2"
)

// UnmarshalLocalFile retrieves JSON or YAML from a file on disk.
// The caller is responsible for checking error return values.
func UnmarshalLocalFile(path string, obj interface{}) error {
	data, err := os.ReadFile(path)
	if err == nil {
		err = unmarshalObject(data, obj)
	}
	return err
}

func unmarshalObject(data []byte, obj interface{}) error {
	return yaml.Unmarshal(data, obj)
}

func Unmarshal(data []byte, obj interface{}) error {
	return unmarshalObject(data, obj)
}

// MarshalLocalYAMLFile writes JSON or YAML to a file on disk.
// The caller is responsible for checking error return values.
func MarshalLocalYAMLFile(path string, obj interface{}) error {
	yamlData, err := yaml.Marshal(obj)
	if err == nil {
		err = os.WriteFile(path, yamlData, 0o600)
	}
	return err
}
