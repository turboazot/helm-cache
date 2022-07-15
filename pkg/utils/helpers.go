package utils

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func WriteStringToFile(path string, content string) error {
	directory := filepath.Dir(path)

	if directory != "." {
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func WriteYamlToFile(in interface{}, path string) error {
	d, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	err = WriteStringToFile(path, string(d))

	return err
}
