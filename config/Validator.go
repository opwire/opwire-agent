package config

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

const RESOURCE_NAME_PATTERN string = `[a-zA-Z][a-zA-Z0-9_-]*`
const BASEURL_PATTERN string = `([\\/]|([\\/][a-zA-Z]|[\\/][a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9])+)?`
const TIMEOUT_PATTERN string = `([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+[uµ]s)?([0-9]+ns)?`
const VERSION_PATTERN string = `[v]?(\\d+\\.)?(\\d+\\.)?(\\*|\\d+)`

type Validator struct {
	schemaLoader gojsonschema.JSONLoader
}

type ValidationResult = gojsonschema.Result
type ValidationError = gojsonschema.ResultError

func NewValidator() (*Validator) {
	validator := &Validator{}
	validator.schemaLoader = gojsonschema.NewStringLoader(configurationSchema)
	return validator
}

func (v *Validator) Validate(cfg *Configuration) (*ValidationResult, error) {
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
			"pattern": "^` + VERSION_PATTERN + `$"
		},
		"agent": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"type": "object",
					"properties": {
						"explanation": {
							"oneOf": [
								{
									"type": "null"
								},
								{
									"$ref": "#/definitions/SectionExplanation"
								}
							]
						},
						"combine-stderr-stdout": {
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
		"logging": {
			"oneOf": [
				{
					"type": "null"
				},
				{
					"$ref": "#/definitions/Logging"
				}
			]
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
				"enabled": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
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
		"Logging": {
			"type": "object",
			"properties": {
				"enabled": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"format": {
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
				"level": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string",
							"enum": [ "debug", "info", "warn", "error", "panic", "fatal" ]
						}
					]
				},
				"output-paths": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "array",
							"items": {
								"type": "string"
							}
						}
					]
				},
				"error-output-paths": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "array",
							"items": {
								"type": "string"
							}
						}
					]
				}
			}
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
				"max-header-bytes": {
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
				"read-timeout": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string",
							"pattern": "^` + TIMEOUT_PATTERN + `$"
						}
					]
				},
				"write-timeout": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string",
							"pattern": "^` + TIMEOUT_PATTERN + `$"
						}
					]
				},
				"concurrent-limit": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"$ref": "#/definitions/sectionConcurrentLimit"
						}
					]
				},
				"single-flight": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"$ref": "#/definitions/sectionSingleFlight"
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
		},
		"SectionExplanation": {
			"type": "object",
			"properties": {
				"enabled": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"format": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				}
			},
			"additionalProperties": false
		},
		"sectionConcurrentLimit": {
			"type": "object",
			"properties": {
				"enabled": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"total": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "integer",
							"minimum": 1
						}
					]
				}
			},
			"additionalProperties": false
		},
		"sectionSingleFlight": {
			"type": "object",
			"properties": {
				"enabled": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"req-id": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				},
				"by-method": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"by-path": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"by-headers": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				},
				"by-queries": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "string"
						}
					]
				},
				"by-body": {
					"oneOf": [
						{
							"type": "null"
						},
						{
							"type": "boolean"
						}
					]
				},
				"by-userip": {
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
