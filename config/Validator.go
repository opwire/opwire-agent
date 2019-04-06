package config

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

const RESOURCE_NAME_PATTERN string = `[a-zA-Z][a-zA-Z0-9_-]*`
const BASEURL_PATTERN string = `([\\/]|([\\/][a-zA-Z]|[\\/][a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9])+)?`

type Validator struct {
	schemaLoader gojsonschema.JSONLoader
}

type ValidationResult interface {
	Valid() bool
	Errors() []gojsonschema.ResultError
}

func NewValidator() (*Validator) {
	validator := &Validator{}
	validator.schemaLoader = gojsonschema.NewStringLoader(configurationSchema)
	return validator
}

func (v *Validator) Validate(cfg *Configuration) (ValidationResult, error) {
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
			"pattern": "^[v]?(\\d+\\.)?(\\d+\\.)?(\\*|\\d+)$"
		},
		"agent": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"type": "object",
					"properties": {
						"explanation-enabled": {
							"oneOf": [
								{
									"type": "null"
								},
								{
									"type": "boolean"
								}
							]
						}
					}
				}
			]
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
						"^` + RESOURCE_NAME_PATTERN + `$": {
							"$ref": "#/definitions/CommandEntrypoint"
						}
					},
					"additionalProperties": false
				}
			]
		},
		"settings": {
			"$ref": "#/definitions/Settings"
		},
		"settings-format": {
			"$ref": "#/definitions/SettingsFormat"
		},
		"http-server": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"$ref": "#/definitions/HttpServer"
				}
			]
		}
	},
	"definitions": {
		"CommandEntrypoint": {
			"type": "object",
			"properties": {
				"default": {
					"$ref": "#/definitions/CommandDescriptor"
				},
				"methods": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "object",
							"patternProperties": {
								"^(?i)(GET|POST|PUT|PATCH|DELETE)$": {
									"$ref": "#/definitions/CommandDescriptor"
								}
							},
							"additionalProperties": false
						}
					]
				},
				"pattern": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string",
							"minLength": 1
						}
					]
				},
				"settings": {
					"$ref": "#/definitions/Settings"
				},
				"settings-format": {
					"$ref": "#/definitions/SettingsFormat"
				}
			}
		},
		"CommandDescriptor": {
			"type": "object",
			"properties": {
				"command": {
					"type": "string"
				},
				"timeout": {
					"type": "number",
					"minimum": 0
				}
			},
			"required": [ "command" ]
		},
		"Settings": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"type": "object"
				}
			]
		},
		"SettingsFormat": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"type": "string",
					"enum": [ "json", "flat" ]
				}
			]
		},
		"HttpServer": {
			"type": "object",
			"properties": {
				"host": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				},
				"port": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "integer",
							"minimum": 0
						}
					]
				},
				"baseurl": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string",
							"pattern": "^` + BASEURL_PATTERN + `$"
						}
					]
				},
				"concurrent-limit-enabled": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"concurrent-limit-total": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "integer"
						}
					]
				},
				"single-flight-enabled": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"single-flight-req-id": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				},
				"single-flight-by-method": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"single-flight-by-path": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"single-flight-by-headers": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				},
				"single-flight-by-queries": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				},
				"single-flight-by-body": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"single-flight-by-userip": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				}
			},
			"additionalProperties": false
		}
	}
}`