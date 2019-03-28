package utils

import (
	"encoding/json"
	"fmt"
)

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
