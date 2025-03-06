package yamlx

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ReadFile(fileName string, structure interface{}) error {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(f, structure)
}
