package cmd

import "fmt"

func schemaContract(resourceType, outputMode string, variants []SchemaOutputVariant, streamEvents []SchemaObjectContract) SchemaOutputContract {
	success := baseEnvelopeContract(resourceType, outputMode, true)
	errorEnvelope := baseEnvelopeContract(resourceType, outputMode, false)
	partial := partialEnvelopeContract(resourceType, outputMode)

	return SchemaOutputContract{
		OutputMode:      outputMode,
		SuccessEnvelope: success,
		ErrorEnvelope:   errorEnvelope,
		PartialEnvelope: &partial,
		Variants:        variants,
		StreamEvents:    streamEvents,
	}
}

func schemaVariant(name, resourceType, outputMode string, data SchemaObjectContract) SchemaOutputVariant {
	envelope := baseEnvelopeContract(resourceType, outputMode, true)
	return SchemaOutputVariant{
		Name:        name,
		Description: variantDescription(name, resourceType),
		OutputMode:  outputMode,
		Envelope:    &envelope,
		Data:        &data,
	}
}

func registryListOutputContract(resourceType, idField string) SchemaOutputContract {
	full := SchemaObjectContract{
		Type:        "array",
		Description: fmt.Sprintf("%s registry entries", resourceType),
		Fields: []SchemaField{
			{Name: idField, Type: "string", Description: fmt.Sprintf("%s identifier", resourceType)},
			{Name: "name", Type: "string", Description: fmt.Sprintf("display name for %s", resourceType)},
			{Name: "attributes", Type: "object", Description: "resource-specific attributes", AdditionalProps: true},
		},
	}
	brief := SchemaObjectContract{
		Type:        "array",
		Description: fmt.Sprintf("brief %s registry rows", resourceType),
		Fields: []SchemaField{
			{Name: idField, Type: "string", Description: fmt.Sprintf("%s identifier", resourceType)},
			{Name: "name", Type: "string", Description: fmt.Sprintf("display name for %s", resourceType)},
		},
	}
	count := SchemaObjectContract{
		Type:        "object",
		Description: fmt.Sprintf("count of %s entries", resourceType),
		Fields: []SchemaField{
			{Name: "count", Type: "number", Required: true, Description: "number of matching items"},
		},
	}

	variants := []SchemaOutputVariant{
		schemaVariant("full", resourceType, "json_envelope", full),
		schemaVariant("brief", resourceType, "json_envelope", brief),
		schemaVariant("count", resourceType, "json_envelope", count),
	}

	return schemaContract(resourceType, "json_envelope", variants, nil)
}

func registryGetOutputContract(resourceType, idField string) SchemaOutputContract {
	full := SchemaObjectContract{
		Type:        "object",
		Description: fmt.Sprintf("single %s registry entry", resourceType),
		Fields: []SchemaField{
			{Name: idField, Type: "string", Description: fmt.Sprintf("%s identifier", resourceType)},
			{Name: "name", Type: "string", Description: fmt.Sprintf("display name for %s", resourceType)},
			{Name: "related", Type: "array", ItemType: "object", Description: "related resources when requested"},
			{Name: "attributes", Type: "object", Description: "resource-specific attributes", AdditionalProps: true},
		},
	}

	return schemaContract(resourceType, "json_envelope", []SchemaOutputVariant{
		schemaVariant("full", resourceType, "json_envelope", full),
	}, nil)
}

func outputContractOverrides() map[string]SchemaOutputContract {
	streamEvent := SchemaObjectContract{
		Type:        "object",
		Description: "ESPHome stream event",
		Fields: []SchemaField{
			{Name: "event", Type: "string", Required: true, Description: "event type such as line or exit"},
			{Name: "data", Type: "string", Description: "stream output line"},
			{Name: "code", Type: "number", Description: "exit code for exit events"},
		},
	}

	return map[string]SchemaOutputContract{
		"hab overview": schemaContract("instance", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "instance", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "home assistant overview summary",
				Fields: []SchemaField{
					{Name: "location_name", Type: "string", Description: "instance location name"},
					{Name: "version", Type: "string", Description: "Home Assistant version"},
					{Name: "state", Type: "string", Description: "instance state"},
					{Name: "time_zone", Type: "string", Description: "configured timezone"},
					{Name: "floors", Type: "number", Description: "floor count"},
					{Name: "areas", Type: "number", Description: "area count"},
					{Name: "devices", Type: "number", Description: "device count"},
					{Name: "entities", Type: "number", Description: "entity count"},
					{Name: "automations", Type: "number", Description: "automation entity count"},
					{Name: "scripts", Type: "number", Description: "script entity count"},
					{Name: "helpers", Type: "number", Description: "helper count"},
					{Name: "entities_by_domain", Type: "object", Description: "domain histogram", AdditionalProps: true},
				},
			}),
		}, nil),
		"hab auth status": schemaContract("auth", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "auth", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "authentication status payload",
				Fields: []SchemaField{
					{Name: "authenticated", Type: "boolean", Required: true, Description: "whether credentials are available"},
					{Name: "url", Type: "string", Description: "Home Assistant base URL"},
					{Name: "auth_type", Type: "string", Description: "auth mechanism in use"},
					{Name: "credential_source", Type: "string", Description: "credential origin"},
					{Name: "token_expiry", Type: "string", Description: "oauth token expiry when available"},
					{Name: "message", Type: "string", Description: "status message when unauthenticated"},
				},
			}),
		}, nil),
		"hab capability probe": schemaContract("capability", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "capability", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "runtime capability probe result",
				Fields: []SchemaField{
					{Name: "auth", Type: "object", Description: "authentication status", AdditionalProps: true},
					{Name: "environment", Type: "object", Description: "detected environment details", AdditionalProps: true},
					{Name: "checks", Type: "object", Description: "detailed capability checks", AdditionalProps: true},
					{Name: "capabilities", Type: "object", Description: "flattened capability booleans", AdditionalProps: true},
				},
			}),
		}, nil),
		"hab entity search": schemaContract("entity", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "entity", "json_envelope", SchemaObjectContract{
				Type:        "array",
				Description: "entity search rows",
				Fields: []SchemaField{
					{Name: "entity_id", Type: "string", Description: "entity identifier"},
					{Name: "name", Type: "string", Description: "friendly name"},
					{Name: "state", Type: "string", Description: "current state"},
				},
			}),
		}, nil),
		"hab action docs": schemaContract("action", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "action", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "action documentation payload",
				Fields: []SchemaField{
					{Name: "action", Type: "string", Description: "domain.action identifier"},
					{Name: "name", Type: "string", Description: "display name"},
					{Name: "description", Type: "string", Description: "human-readable description"},
					{Name: "fields", Type: "object", Description: "supported data fields", AdditionalProps: true},
					{Name: "target", Type: "object", Description: "target selector schema", AdditionalProps: true},
				},
			}),
		}, nil),
		"hab dashboard list": schemaContract("dashboard", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "dashboard", "json_envelope", SchemaObjectContract{
				Type:        "array",
				Description: "dashboard list rows",
				Fields: []SchemaField{
					{Name: "url_path", Type: "string", Description: "dashboard URL path"},
					{Name: "title", Type: "string", Description: "dashboard title"},
					{Name: "mode", Type: "string", Description: "dashboard mode"},
					{Name: "show_in_sidebar", Type: "boolean", Description: "sidebar visibility"},
					{Name: "require_admin", Type: "boolean", Description: "admin requirement"},
				},
			}),
			schemaVariant("brief", "dashboard", "json_envelope", SchemaObjectContract{
				Type:        "array",
				Description: "brief dashboard rows",
				Fields: []SchemaField{
					{Name: "url_path", Type: "string", Description: "dashboard URL path"},
					{Name: "title", Type: "string", Description: "dashboard title"},
				},
			}),
			schemaVariant("count", "dashboard", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "dashboard count payload",
				Fields:      []SchemaField{{Name: "count", Type: "number", Required: true, Description: "number of dashboards"}},
			}),
		}, nil),
		"hab dashboard get": schemaContract("dashboard", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "dashboard", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "dashboard configuration payload",
				Fields: []SchemaField{
					{Name: "title", Type: "string", Description: "dashboard title"},
					{Name: "views", Type: "array", ItemType: "object", Description: "dashboard views"},
					{Name: "config", Type: "object", Description: "full dashboard config", AdditionalProps: true},
				},
			}),
		}, nil),
		"hab calendar list": schemaContract("calendar_event", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "calendar_event", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "calendar event list wrapper",
				Fields: []SchemaField{
					{Name: "events", Type: "array", ItemType: "object", Description: "calendar events in requested range"},
				},
			}),
		}, nil),
		"hab todo items": schemaContract("todo_item", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "todo_item", "json_envelope", SchemaObjectContract{
				Type:        "array",
				Description: "to-do list items",
				Fields: []SchemaField{
					{Name: "uid", Type: "string", Description: "to-do item identifier"},
					{Name: "summary", Type: "string", Description: "item summary"},
					{Name: "status", Type: "string", Description: "completion status"},
					{Name: "description", Type: "string", Description: "item description"},
					{Name: "due", Type: "string", Description: "due date or datetime"},
				},
			}),
		}, nil),
		"hab esphome update": schemaContract("esphome", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "esphome", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "ESPhome update workflow result",
				Fields: []SchemaField{
					{Name: "configuration", Type: "string", Required: true, Description: "target configuration filename"},
					{Name: "validated", Type: "boolean", Required: true, Description: "validation step completed"},
					{Name: "written", Type: "boolean", Required: true, Description: "configuration file changed"},
					{Name: "build", Type: "boolean", Required: true, Description: "build step completed"},
					{Name: "upload", Type: "boolean", Required: true, Description: "upload step completed"},
					{Name: "run", Type: "boolean", Required: true, Description: "combined run step completed"},
					{Name: "build_events", Type: "array", ItemType: "object", Description: "build stream events when json mode is used"},
					{Name: "upload_events", Type: "array", ItemType: "object", Description: "upload stream events when json mode is used"},
					{Name: "run_events", Type: "array", ItemType: "object", Description: "run stream events when json mode is used"},
				},
			}),
			schemaVariant("plan", "esphome", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "execution plan for esphome update",
				Fields:      []SchemaField{{Name: "mode", Type: "string", Enum: []string{"plan"}, Required: true}, {Name: "would_change", Type: "boolean", Required: true}, {Name: "target", Type: "object", Required: true, AdditionalProps: true}, {Name: "inputs", Type: "object", AdditionalProps: true}, {Name: "steps", Type: "array", ItemType: "string", Required: true}, {Name: "verification_commands", Type: "array", ItemType: "string"}},
			}),
		}, nil),
		"hab esphome validate": schemaContract("esphome", "json_envelope", []SchemaOutputVariant{
			schemaVariant("full", "esphome", "json_envelope", SchemaObjectContract{
				Type:        "object",
				Description: "structured validation payload when --structured is used",
				Fields: []SchemaField{
					{Name: "configuration", Type: "string", Description: "validated configuration"},
					{Name: "parsed", Type: "object", Description: "parsed ESPHome config", AdditionalProps: true},
				},
			}),
			schemaVariant("stream", "esphome", "ndjson_stream", SchemaObjectContract{
				Type:        "array",
				Description: "native validator stream events",
				Fields:      streamEvent.Fields,
			}),
		}, []SchemaObjectContract{streamEvent}),
		"hab esphome logs": schemaContract("esphome", "ndjson_stream", []SchemaOutputVariant{
			schemaVariant("stream", "esphome", "ndjson_stream", SchemaObjectContract{
				Type:        "array",
				Description: "ESPHome log stream events",
				Fields:      streamEvent.Fields,
			}),
		}, []SchemaObjectContract{streamEvent}),
	}
}
