package config

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

type ConfigValidator struct {
	schemaLoader gojsonschema.JSONLoader
}

type ValidationResult interface {
	Valid() bool
	Errors() []gojsonschema.ResultError
}

func NewValidator() (*ConfigValidator) {
	validator := &ConfigValidator{}
	validator.schemaLoader = gojsonschema.NewStringLoader(configurationSchema)
	return validator
}

func (v *ConfigValidator) Validate(cfg *Configuration) (ValidationResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("The configuration object is nil")
	}
	if v.schemaLoader == nil {
		return nil, fmt.Errorf("Validator is not initialized properly")
	}
	documentLoader := gojsonschema.NewGoLoader(cfg)
	return gojsonschema.Validate(v.schemaLoader, documentLoader)
}

var configurationSchema string = `{
	"type": "object",
	"properties": {
		"version": {
			"type": "string",
			"pattern": "^(\\d+\\.)?(\\d+\\.)?(\\*|\\d+)$"
		},
		"main-resource": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"$ref": "#/definitions/CommandEntrypoint"
				}
			]
		},
		"resources": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"type": "object",
					"patternProperties": {
						"^[a-zA-Z][a-zA-Z0-9_-]*$": {
							"$ref": "#/definitions/CommandEntrypoint"
						}
					},
					"additionalProperties": false
				}
			]
		}
	},
	"definitions": {
		"CommandEntrypoint": {
			"type": "object",
			"properties": {
				"default": {
					"type": "object"
				},
				"methods": {
					"type": "object",
					"patternProperties": {
						"^GET|POST|PUT|PATCH|DELETE|[a-zA-Z]+$": {
							"type": "object"
						}
					},
					"additionalProperties": false
				}
			}
		}
	}
}`