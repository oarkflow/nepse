package utils

import (
	"encoding/json"
	"os"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func ReadJson[T any](data []byte) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return v, err
}

func ReadJsonFile[T any](file string) (T, error) {
	var v T
	content, err := os.ReadFile(file)
	if err != nil {
		return v, err
	}
	return ReadJson[T](content)
}

func WriteJsonFile(file string, data any) error {
	dataByte, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(file, dataByte, os.ModePerm)
}
