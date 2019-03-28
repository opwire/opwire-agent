package utils

import (
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
)

func CombineSettings(resourceSettings map[string]interface{}, sharedSettings map[string]interface{}) (map[string]interface{}, error) {
	mySettings := make(map[string]interface{}, 0)
	if resourceSettings != nil {
		if err := mergo.Merge(&mySettings, resourceSettings); err != nil {
			return mySettings, err
		}
	}
	if sharedSettings != nil {
		if err := mergo.Merge(&mySettings, sharedSettings); err != nil {
			return mySettings, err
		}
	}
	return mySettings, nil
}

func TransformSettingsToEnvs(prefix string, settings map[string]interface{}, format string) ([]string, error) {
	var envs []string
	var err error
	if format == "flat" {
		envs, err = convertSettingsToFlatEnvs(prefix, settings)
	} else {
		envs, err = convertSettingsToJsonEnvs(prefix, settings)
	}
	if err != nil {
		return nil, err
	}
	return envs, nil
}

func convertSettingsToJsonEnvs(prefix string, settings map[string]interface{}) ([]string, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	return []string{fmt.Sprintf("%s=%s", prefix, data)}, nil
}

func convertSettingsToFlatEnvs(prefix string, settings map[string]interface{}) ([]string, error) {
	pairs, err := Flatten(prefix, settings)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(pairs))
	for key, val := range pairs {
		if val != nil {
			item := fmt.Sprintf("%s=%v", key, val)
			list = append(list, item)
		}
	}
	return list, nil
}
