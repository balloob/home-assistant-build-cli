// Package client provides REST, WebSocket, and ESPHome API clients for
// communicating with Home Assistant. REST is used for state queries and
// service calls; WebSocket is used for registry operations and subscriptions.
package client

// Compile-time assertions that concrete types implement the interfaces.
var (
	_ WebSocketAPI = (*WebSocketClient)(nil)
	_ RestAPI      = (*RestClient)(nil)
	_ ESPHomeAPI   = (*ESPHomeClient)(nil)
)

// Compile-time assertions that WebSocketClient satisfies each sub-interface.
var (
	_ WebSocketConnection = (*WebSocketClient)(nil)
	_ WebSocketCommander  = (*WebSocketClient)(nil)
	_ StateAPI            = (*WebSocketClient)(nil)
	_ AreaRegistryAPI     = (*WebSocketClient)(nil)
	_ FloorRegistryAPI    = (*WebSocketClient)(nil)
	_ LabelRegistryAPI    = (*WebSocketClient)(nil)
	_ DeviceRegistryAPI   = (*WebSocketClient)(nil)
	_ EntityRegistryAPI   = (*WebSocketClient)(nil)
	_ ZoneAPI             = (*WebSocketClient)(nil)
	_ SystemHealthAPI     = (*WebSocketClient)(nil)
	_ SearchAPI           = (*WebSocketClient)(nil)
	_ HelperAPI           = (*WebSocketClient)(nil)
	_ ConfigAPI           = (*WebSocketClient)(nil)
	_ PersonRegistryAPI   = (*WebSocketClient)(nil)
	_ CategoryRegistryAPI = (*WebSocketClient)(nil)
	_ IntegrationAPI      = (*WebSocketClient)(nil)
	_ TodoAPI             = (*WebSocketClient)(nil)
	_ NotificationAPI     = (*WebSocketClient)(nil)
	_ RepairAPI           = (*WebSocketClient)(nil)
)

// ---------------------------------------------------------------------------
// Domain-specific WebSocket interfaces
//
// Each interface captures a cohesive set of operations for a single domain.
// Commands that only need one domain can accept the narrow interface, making
// dependencies explicit and mocks trivial.
// ---------------------------------------------------------------------------

// WebSocketConnection handles the WebSocket connection lifecycle.
type WebSocketConnection interface {
	Connect() error
	Close() error
}

// WebSocketCommander provides the generic SendCommand escape hatch for
// domains that do not yet have dedicated typed methods (dashboards, backups,
// blueprints, threads, traces, etc.).
type WebSocketCommander interface {
	SendCommand(cmdType string, params map[string]interface{}) (interface{}, error)
}

// StateAPI provides state and service queries.
type StateAPI interface {
	GetStates() ([]interface{}, error)
	GetConfig() (map[string]interface{}, error)
	GetServices() (map[string]interface{}, error)
	CallService(domain, service string, data, target map[string]interface{}, returnResponse bool) (interface{}, error)
	Ping() error
}

// AreaRegistryAPI provides CRUD operations on the area registry.
type AreaRegistryAPI interface {
	AreaRegistryList() ([]interface{}, error)
	AreaRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error)
	AreaRegistryUpdate(areaID string, params map[string]interface{}) (map[string]interface{}, error)
	AreaRegistryDelete(areaID string) error
}

// FloorRegistryAPI provides CRUD operations on the floor registry.
type FloorRegistryAPI interface {
	FloorRegistryList() ([]interface{}, error)
	FloorRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error)
	FloorRegistryUpdate(floorID string, params map[string]interface{}) (map[string]interface{}, error)
	FloorRegistryDelete(floorID string) error
}

// LabelRegistryAPI provides CRUD operations on the label registry.
type LabelRegistryAPI interface {
	LabelRegistryList() ([]interface{}, error)
	LabelRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error)
	LabelRegistryUpdate(labelID string, params map[string]interface{}) (map[string]interface{}, error)
	LabelRegistryDelete(labelID string) error
}

// DeviceRegistryAPI provides operations on the device registry.
type DeviceRegistryAPI interface {
	DeviceRegistryList() ([]interface{}, error)
	DeviceRegistryUpdate(deviceID string, params map[string]interface{}) (map[string]interface{}, error)
}

// EntityRegistryAPI provides operations on the entity registry.
type EntityRegistryAPI interface {
	EntityRegistryList() ([]interface{}, error)
	EntityRegistryGet(entityID string) (map[string]interface{}, error)
	EntityRegistryUpdate(entityID string, params map[string]interface{}) (map[string]interface{}, error)
}

// ZoneAPI provides CRUD operations on zones.
type ZoneAPI interface {
	ZoneList() ([]interface{}, error)
	ZoneCreate(name string, latitude, longitude, radius float64, params map[string]interface{}) (map[string]interface{}, error)
	ZoneUpdate(zoneID string, params map[string]interface{}) (map[string]interface{}, error)
	ZoneDelete(zoneID string) error
}

// SystemHealthAPI provides system health information.
type SystemHealthAPI interface {
	SystemHealthInfo() (map[string]interface{}, error)
}

// SearchAPI provides cross-domain search operations.
type SearchAPI interface {
	SearchRelated(itemType, itemID string) (map[string][]string, error)
}

// HelperAPI provides CRUD operations on WebSocket-based helpers
// (input_boolean, counter, timer, schedule, etc.).
type HelperAPI interface {
	HelperList(helperType string) ([]interface{}, error)
	HelperCreate(helperType string, params map[string]interface{}) (map[string]interface{}, error)
	HelperUpdate(helperType, helperID string, params map[string]interface{}) (map[string]interface{}, error)
	HelperDelete(helperType, helperID string) error
}

// PersonRegistryAPI provides CRUD operations on the person registry.
type PersonRegistryAPI interface {
	PersonRegistryList() ([]interface{}, error)
	PersonRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error)
	PersonRegistryUpdate(personID string, params map[string]interface{}) (map[string]interface{}, error)
	PersonRegistryDelete(personID string) error
}

// CategoryRegistryAPI provides CRUD operations on the category registry.
// Categories are scoped per table (e.g. "automation", "script", "scene", "helpers").
type CategoryRegistryAPI interface {
	CategoryRegistryList(scope string) ([]interface{}, error)
	CategoryRegistryCreate(scope, name string, params map[string]interface{}) (map[string]interface{}, error)
	CategoryRegistryUpdate(categoryID string, params map[string]interface{}) (map[string]interface{}, error)
	CategoryRegistryDelete(scope, categoryID string) error
}

// ConfigAPI provides config flow and config entry operations.
type ConfigAPI interface {
	ConfigFlowInit(handler string, context map[string]interface{}) (map[string]interface{}, error)
	ConfigFlowConfigure(flowID string, data map[string]interface{}) (map[string]interface{}, error)
	ConfigEntriesList(domain string) ([]interface{}, error)
	ConfigEntryDelete(entryID string) error
	ResolveEntityToConfigEntry(entityID string) (string, error)
}

// IntegrationAPI provides management operations on config entries (integrations).
// It extends the basic ConfigEntriesList/ConfigEntryDelete on ConfigAPI with
// single-entry lookup, enable/disable, and reload.
type IntegrationAPI interface {
	ConfigEntryGet(entryID string) (map[string]interface{}, error)
	ConfigEntrySetDisabled(entryID string, disabledBy interface{}) (map[string]interface{}, error)
}

// TodoAPI provides read access to to-do list items.
// Item mutations go through the REST service call API (todo.add_item, etc.).
type TodoAPI interface {
	TodoItemList(entityID string) ([]interface{}, error)
}

// NotificationAPI provides read access to persistent notifications.
// Create/dismiss go through the REST service call API.
type NotificationAPI interface {
	NotificationList() ([]interface{}, error)
}

// RepairAPI provides access to the HA repairs/issues system.
type RepairAPI interface {
	RepairListIssues() ([]interface{}, error)
	RepairIgnoreIssue(domain, issueID string, ignore bool) error
}

// ---------------------------------------------------------------------------
// Composed interfaces
// ---------------------------------------------------------------------------

// WebSocketAPI defines the full interface for WebSocket operations against
// Home Assistant. It is composed from the domain-specific interfaces above.
//
// Commands that need the full surface area (e.g., overview) accept this type.
// Commands that need a single domain can accept the narrower interface instead,
// making dependencies explicit and mocks trivial to implement.
type WebSocketAPI interface {
	WebSocketConnection
	WebSocketCommander
	StateAPI
	AreaRegistryAPI
	FloorRegistryAPI
	LabelRegistryAPI
	DeviceRegistryAPI
	EntityRegistryAPI
	ZoneAPI
	SystemHealthAPI
	SearchAPI
	HelperAPI
	ConfigAPI
	PersonRegistryAPI
	CategoryRegistryAPI
	IntegrationAPI
	TodoAPI
	NotificationAPI
	RepairAPI
}

// ESPHomeAPI defines the interface for operations against the ESPHome Dashboard.
// This enables unit-testing ESPHome command handlers with mock implementations.
type ESPHomeAPI interface {
	GetDevices() (*ESPHomeDeviceList, error)
	GetPing() (map[string]*bool, error)
	GetVersion() (string, error)
	ReadConfig(configuration string) (string, error)
	WriteConfig(configuration, content string) error
	StreamCommand(path string, spawnMsg map[string]interface{}, callback func(ESPHomeStreamEvent)) (int, error)
}

// RestAPI defines the interface for REST operations against Home Assistant.
// This enables unit-testing command handlers with mock implementations.
//
// Some operations (GetConfig, GetStates, GetServices, ConfigEntryDelete) are
// also available on WebSocketAPI. REST variants are used when WebSocket is
// unnecessary or unavailable; WebSocket variants are used when the caller
// already holds a connection or needs subscription semantics.
type RestAPI interface {
	// Generic HTTP methods
	Get(endpoint string) (interface{}, error)
	Post(endpoint string, body interface{}) (interface{}, error)
	Put(endpoint string, body interface{}) (interface{}, error)
	Delete(endpoint string) (interface{}, error)

	// High-level API methods
	GetConfig() (map[string]interface{}, error)
	GetStates() ([]interface{}, error)
	GetState(entityID string) (map[string]interface{}, error)
	GetServices() ([]interface{}, error)
	CallService(domain, service string, data map[string]interface{}) (interface{}, error)
	CheckConfig() (map[string]interface{}, error)
	Restart() error
	GetErrorLog() (string, error)
	GetHistory(entityID string, startTime, endTime string) ([]interface{}, error)
	GetLogbook(entityID string, startTime, endTime string) ([]interface{}, error)
	RenderTemplate(template string) (string, error)
	GetEvents() ([]interface{}, error)
	FireEvent(eventType string, data map[string]interface{}) error
	ConfigEntryReload(entryID string) (map[string]interface{}, error)

	// Config flow methods
	ConfigFlowCreate(handler string) (map[string]interface{}, error)
	ConfigFlowStep(flowID string, data map[string]interface{}) (map[string]interface{}, error)
	ConfigEntryDelete(entryID string) error
}
