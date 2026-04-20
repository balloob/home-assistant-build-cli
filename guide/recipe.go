package guide

// RecipeInput defines a named input required by a recipe.
type RecipeInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Source      string `json:"source,omitempty"`
}

// RecipeCapture describes a value extracted from a step result for later use.
type RecipeCapture struct {
	Field       string `json:"field"`
	As          string `json:"as"`
	Description string `json:"description,omitempty"`
}

// RecipeStep defines an executable step in a guide workflow recipe.
type RecipeStep struct {
	ID              string          `json:"id"`
	Type            string          `json:"type"`
	Summary         string          `json:"summary"`
	CommandTemplate string          `json:"command_template,omitempty"`
	CommandPath     string          `json:"command_path,omitempty"`
	Expects         []string        `json:"expects,omitempty"`
	Captures        []RecipeCapture `json:"captures,omitempty"`
	OnFailure       []string        `json:"on_failure,omitempty"`
	Notes           []string        `json:"notes,omitempty"`
}

// RecipeBranch describes simple workflow branching conditions.
type RecipeBranch struct {
	When   string `json:"when"`
	StepID string `json:"step_id"`
}

// Recipe describes a machine-readable workflow that an LLM can execute.
type Recipe struct {
	ID                   string         `json:"id"`
	Summary              string         `json:"summary"`
	RequiredCapabilities []string       `json:"required_capabilities,omitempty"`
	Inputs               []RecipeInput  `json:"inputs,omitempty"`
	Steps                []RecipeStep   `json:"steps"`
	Branches             []RecipeBranch `json:"branches,omitempty"`
	VerificationSteps    []string       `json:"verification_steps,omitempty"`
	RecoverySteps        []string       `json:"recovery_steps,omitempty"`
}

func recipesForTopic(id string) []Recipe {
	switch id {
	case "index":
		return []Recipe{{
			ID:                   "bootstrap-discovery",
			Summary:              "Authenticate if needed, probe capabilities, and gather high-level discovery context.",
			RequiredCapabilities: []string{"local"},
			Steps: []RecipeStep{
				{ID: "check-auth", Type: "discovery", Summary: "Inspect auth status", CommandTemplate: "hab auth status --json", Expects: []string{"data.authenticated or error.code == AUTH_REQUIRED"}, OnFailure: []string{"login"}},
				{ID: "login", Type: "recovery", Summary: "Authenticate when required", CommandTemplate: "hab auth login", Notes: []string{"Skip this step if auth status already reports authenticated."}},
				{ID: "probe", Type: "discovery", Summary: "Probe runtime capabilities", CommandTemplate: "hab capability probe --json", Expects: []string{"data.capabilities.auth != null"}},
				{ID: "overview", Type: "discovery", Summary: "Collect high-level instance overview", CommandTemplate: "hab overview --json", Expects: []string{"data.entities != null"}},
			},
			VerificationSteps: []string{"probe", "overview"},
			RecoverySteps:     []string{"login"},
		}}
	case "discovery":
		return []Recipe{{
			ID:                   "discover-targets",
			Summary:              "Find target resources and resolve exact IDs before any mutation workflow.",
			RequiredCapabilities: []string{"auth"},
			Inputs:               []RecipeInput{{Name: "search_term", Description: "Entity, area, or device hint to search for", Required: true, Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "overview", Type: "discovery", Summary: "Inspect global instance state", CommandTemplate: "hab overview --json"},
				{ID: "entity-search", Type: "discovery", Summary: "Search entities using the provided hint", CommandTemplate: "hab entity search {{search_term}} --json", Captures: []RecipeCapture{{Field: "data[0].entity_id", As: "entity_id", Description: "candidate entity for follow-up inspection"}}},
				{ID: "related", Type: "verification", Summary: "Inspect relationships for the selected entity", CommandTemplate: "hab search related entity {{entity_id}} --json", Expects: []string{"data != null"}},
			},
			VerificationSteps: []string{"related"},
		}}
	case "auth":
		return []Recipe{{
			ID:                   "authenticate-session",
			Summary:              "Establish and verify an authenticated Home Assistant session.",
			RequiredCapabilities: []string{"local"},
			Steps: []RecipeStep{
				{ID: "status", Type: "discovery", Summary: "Check current auth state", CommandTemplate: "hab auth status --json"},
				{ID: "login", Type: "mutation", Summary: "Run login flow if needed", CommandTemplate: "hab auth login", OnFailure: []string{"refresh"}},
				{ID: "refresh", Type: "recovery", Summary: "Refresh an expired session", CommandTemplate: "hab auth refresh --json"},
				{ID: "verify", Type: "verification", Summary: "Verify authenticated state", CommandTemplate: "hab auth status --json", Expects: []string{"data.authenticated == true"}},
			},
			VerificationSteps: []string{"verify"},
			RecoverySteps:     []string{"refresh"},
		}}
	case "input-output":
		return []Recipe{{
			ID:                   "inspect-command-contract",
			Summary:              "Inspect invocation rules and output contracts before writing payloads.",
			RequiredCapabilities: []string{"local"},
			Inputs:               []RecipeInput{{Name: "command_path", Description: "Command path to inspect", Required: true, Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "schema", Type: "discovery", Summary: "Inspect command schema and output contract", CommandTemplate: "hab schema {{command_path}} --json"},
				{ID: "docs", Type: "verification", Summary: "Inspect action schema for service calls when relevant", CommandTemplate: "hab action docs {{command_path}} --json", Notes: []string{"Only use when the target command is an action/service workflow."}},
			},
			VerificationSteps: []string{"schema"},
		}}
	case "registry":
		return []Recipe{{
			ID:                   "registry-crud",
			Summary:              "Discover a registry resource, mutate it, and verify the resulting ID/state.",
			RequiredCapabilities: []string{"auth", "ws"},
			Inputs:               []RecipeInput{{Name: "resource", Description: "Registry resource family such as area, floor, or label", Required: true, Source: "prompt"}, {Name: "name", Description: "Name for create workflows", Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "list", Type: "discovery", Summary: "List current registry resources", CommandTemplate: "hab {{resource}} list --json", CommandPath: "hab area list"},
				{ID: "create-plan", Type: "discovery", Summary: "Preview the create mutation", CommandTemplate: "hab {{resource}} create {{name}} --plan --json", CommandPath: "hab area create", Expects: []string{"data.mode == plan"}},
				{ID: "create", Type: "mutation", Summary: "Create the resource", CommandTemplate: "hab {{resource}} create {{name}} --json", CommandPath: "hab area create", Captures: []RecipeCapture{{Field: "data.id", As: "resource_id", Description: "resource identifier when present"}, {Field: "data.area_id", As: "resource_id", Description: "resource identifier for area-style resources"}}},
				{ID: "verify", Type: "verification", Summary: "Fetch the created resource", CommandTemplate: "hab {{resource}} get {{resource_id}} --json", CommandPath: "hab area get"},
			},
			VerificationSteps: []string{"verify"},
		}}
	case "automation":
		return []Recipe{{
			ID:                   "automation-write-loop",
			Summary:              "Inspect action schema, create/update an automation or script, then verify it.",
			RequiredCapabilities: []string{"auth"},
			Inputs:               []RecipeInput{{Name: "entity_id", Description: "Entity target for action discovery", Required: true, Source: "prompt"}, {Name: "automation_id", Description: "Automation or script identifier", Required: true, Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "action-docs", Type: "discovery", Summary: "Inspect action/service contract", CommandTemplate: "hab action docs {{entity_id}} --json", Notes: []string{"Substitute a service name such as light.turn_on when known."}},
				{ID: "plan", Type: "discovery", Summary: "Preview automation mutation", CommandTemplate: "hab automation update {{automation_id}} --plan --json"},
				{ID: "apply", Type: "mutation", Summary: "Apply automation or script change", CommandTemplate: "hab automation update {{automation_id}} -f automation.yaml --json", OnFailure: []string{"verify-script"}},
				{ID: "verify", Type: "verification", Summary: "Verify automation state", CommandTemplate: "hab automation get {{automation_id}} --json"},
				{ID: "verify-script", Type: "verification", Summary: "Alternative verification for script workflows", CommandTemplate: "hab script get {{automation_id}} --json"},
			},
			VerificationSteps: []string{"verify", "verify-script"},
		}}
	case "dashboard":
		return []Recipe{{
			ID:                   "dashboard-edit-loop",
			Summary:              "Inspect dashboard structure, plan a mutation, apply it, and verify the updated view/card structure.",
			RequiredCapabilities: []string{"auth", "ws"},
			Inputs:               []RecipeInput{{Name: "dashboard_id", Description: "Dashboard URL path", Required: true, Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "dashboard-get", Type: "discovery", Summary: "Fetch dashboard configuration", CommandTemplate: "hab dashboard get {{dashboard_id}} --json"},
				{ID: "view-list", Type: "discovery", Summary: "List dashboard views", CommandTemplate: "hab dashboard view list {{dashboard_id}} --json"},
				{ID: "card-plan", Type: "discovery", Summary: "Preview card mutation", CommandTemplate: "hab dashboard card create {{dashboard_id}} 0 --plan --json"},
				{ID: "card-apply", Type: "mutation", Summary: "Apply card or view update", CommandTemplate: "hab dashboard card create {{dashboard_id}} 0 -d '{\"type\":\"markdown\",\"content\":\"Hello\"}' --json"},
				{ID: "verify", Type: "verification", Summary: "Verify updated dashboard content", CommandTemplate: "hab dashboard card list {{dashboard_id}} 0 --json"},
			},
			VerificationSteps: []string{"verify"},
		}}
	case "helpers":
		return []Recipe{{
			ID:                   "helper-create-verify",
			Summary:              "List helper types, create a helper, and verify the resulting entity ID.",
			RequiredCapabilities: []string{"auth"},
			Inputs:               []RecipeInput{{Name: "helper_type", Description: "Helper subtype command, for example input-boolean", Required: true, Source: "prompt"}, {Name: "name", Description: "Display name for the helper", Required: true, Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "types", Type: "discovery", Summary: "List available helper types", CommandTemplate: "hab helper types --json"},
				{ID: "create-plan", Type: "discovery", Summary: "Preview helper creation", CommandTemplate: "hab helper {{helper_type}} create {{name}} --plan --json", CommandPath: "hab helper input-boolean create"},
				{ID: "create", Type: "mutation", Summary: "Create the helper", CommandTemplate: "hab helper {{helper_type}} create {{name}} --json", CommandPath: "hab helper input-boolean create", Captures: []RecipeCapture{{Field: "data.id", As: "entity_id", Description: "helper entity identifier"}}},
				{ID: "verify", Type: "verification", Summary: "Verify helper presence", CommandTemplate: "hab entity get {{entity_id}} --json"},
			},
			VerificationSteps: []string{"verify"},
		}}
	case "calendar-todo":
		return []Recipe{{
			ID:                   "todo-item-loop",
			Summary:              "List to-do lists, add an item, and verify the item list afterwards.",
			RequiredCapabilities: []string{"auth"},
			Inputs:               []RecipeInput{{Name: "todo_entity_id", Description: "To-do list entity ID", Required: true, Source: "prompt"}, {Name: "summary", Description: "To-do item summary", Required: true, Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "lists", Type: "discovery", Summary: "List to-do lists", CommandTemplate: "hab todo lists --json"},
				{ID: "items", Type: "discovery", Summary: "List existing to-do items", CommandTemplate: "hab todo items {{todo_entity_id}} --json"},
				{ID: "add", Type: "mutation", Summary: "Add a new to-do item", CommandTemplate: "hab todo add {{todo_entity_id}} --summary '{{summary}}' --json"},
				{ID: "verify", Type: "verification", Summary: "Verify updated to-do items", CommandTemplate: "hab todo items {{todo_entity_id}} --json"},
			},
			VerificationSteps: []string{"verify"},
		}}
	case "esphome":
		return []Recipe{{
			ID:                   "esphome-validate-build",
			Summary:              "Select an ESPHome configuration, validate it, and then build or inspect it.",
			RequiredCapabilities: []string{"auth", "esphome"},
			Inputs:               []RecipeInput{{Name: "configuration", Description: "ESPHome configuration name", Required: true, Source: "prompt"}},
			Steps: []RecipeStep{
				{ID: "list", Type: "discovery", Summary: "List available ESPHome configurations", CommandTemplate: "hab esphome list --json"},
				{ID: "use-context", Type: "mutation", Summary: "Store configuration in ESPHome context", CommandTemplate: "hab esphome context use {{configuration}}"},
				{ID: "validate", Type: "verification", Summary: "Validate the configuration", CommandTemplate: "hab esphome validate {{configuration}}"},
				{ID: "info", Type: "verification", Summary: "Inspect configuration metadata", CommandTemplate: "hab esphome info {{configuration}} --json"},
			},
			VerificationSteps: []string{"validate", "info"},
		}}
	case "operations":
		return []Recipe{{
			ID:                   "operations-health-loop",
			Summary:              "Check health, inspect repairs/diagnostics, and verify system state before and after operational changes.",
			RequiredCapabilities: []string{"auth"},
			Steps: []RecipeStep{
				{ID: "health", Type: "discovery", Summary: "Inspect system health", CommandTemplate: "hab system health --json"},
				{ID: "repairs", Type: "discovery", Summary: "List repair issues", CommandTemplate: "hab repairs list --json"},
				{ID: "diagnostics", Type: "discovery", Summary: "List diagnostics domains", CommandTemplate: "hab diagnostics list --json"},
				{ID: "backup-plan", Type: "discovery", Summary: "Preview a backup or maintenance mutation", CommandTemplate: "hab backup create --plan --json", Notes: []string{"Skip when the environment does not support backups."}},
				{ID: "verify", Type: "verification", Summary: "Re-check health after operations", CommandTemplate: "hab system health --json"},
			},
			VerificationSteps: []string{"verify"},
		}}
	default:
		return nil
	}
}
