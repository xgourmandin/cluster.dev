package utils

import (
	"encoding/json"
	"fmt"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
)

func ParseSecretData(secretData string, secretName string) (interface{}, error) {
	parseCheck := []interface{}{}
	errSliceCheck := json.Unmarshal([]byte(secretData), &parseCheck)
	parsed := map[string]interface{}{}
	err := json.Unmarshal([]byte(secretData), &parsed)
	if err != nil {
		if errSliceCheck != nil {
			log.Debugf("Secret '%v' is not JSON, creating raw data", secretName)
			return secretData, nil
		}
		return nil, fmt.Errorf("get secret: JSON secret must be a map, not array")
	}
	return parsed, nil
}

func MarshallSecretData(secretData interface{}) (string, error) {
	kind := reflect.TypeOf(secretData).Kind()
	var secretDataStr string

	if kind == reflect.Map {
		secretDataByte, err := json.Marshal(secretData)
		if err != nil {
			return "", err
		}
		secretDataStr = string(secretDataByte)
	} else {
		secretDataStr = fmt.Sprintf("%v", secretData)
	}

	if kind == reflect.Slice {
		return "", fmt.Errorf("create secret: array is not allowed")
	}
	return secretDataStr, nil
}

func SaveTmplToFile(name string, data []byte) (string, error) {
	filenameCheck := filepath.Join(config.Global.WorkingDir, name)
	if _, err := os.Stat(filenameCheck); os.IsNotExist(err) {
		err = os.WriteFile(filenameCheck, data, fs.ModePerm)
		if err != nil {
			return "", err
		}
		return filenameCheck, nil
	}
	f, err := os.CreateTemp(config.Global.WorkingDir, "*_"+name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}
