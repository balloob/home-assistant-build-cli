package client

// Compile-time assertions that concrete types implement the interfaces.
var (
	_ WebSocketAPI = (*WebSocketClient)(nil)
	_ RestAPI      = (*RestClient)(nil)
)

// WebSocketAPI defines the interface for WebSocket operations against Home Assistant.
// This enables unit-testing command handlers with mock implementations.
type WebSocketAPI interface {
	// Connection lifecycle
	Connect() error
	Close() error

	// Generic command
	SendCommand(cmdType string, params map[string]interface{}) (interface{}, error)

	// State queries
	GetStates() ([]interface{}, error)
	GetConfig() (map[string]interface{}, error)
	GetServices() (map[string]interface{}, error)
	CallService(domain, service string, data, target map[string]interface{}, returnResponse bool) (interface{}, error)
	Ping() error

	// Area registry
	AreaRegistryList() ([]interface{}, error)
	AreaRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error)
	AreaRegistryUpdate(areaID string, params map[string]interface{}) (map[string]interface{}, error)
	AreaRegistryDelete(areaID string) error

	// Floor registry
	FloorRegistryList() ([]interface{}, error)
	FloorRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error)
	FloorRegistryUpdate(floorID string, params map[string]interface{}) (map[string]interface{}, error)
	FloorRegistryDelete(floorID string) error

	// Label registry
	LabelRegistryList() ([]interface{}, error)
	LabelRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error)
	LabelRegistryUpdate(labelID string, params map[string]interface{}) (map[string]interface{}, error)
	LabelRegistryDelete(labelID string) error

	// Device registry
	DeviceRegistryList() ([]interface{}, error)
	DeviceRegistryUpdate(deviceID string, params map[string]interface{}) (map[string]interface{}, error)

	// Entity registry
	EntityRegistryList() ([]interface{}, error)
	EntityRegistryGet(entityID string) (map[string]interface{}, error)
	EntityRegistryUpdate(entityID string, params map[string]interface{}) (map[string]interface{}, error)

	// Zones
	ZoneList() ([]interface{}, error)
	ZoneCreate(name string, latitude, longitude, radius float64, params map[string]interface{}) (map[string]interface{}, error)
	ZoneUpdate(zoneID string, params map[string]interface{}) (map[string]interface{}, error)
	ZoneDelete(zoneID string) error

	// System
	SystemHealthInfo() (map[string]interface{}, error)

	// Search
	SearchRelated(itemType, itemID string) (map[string][]string, error)

	// Helpers (WS-based)
	HelperList(helperType string) ([]interface{}, error)
	HelperCreate(helperType string, params map[string]interface{}) (map[string]interface{}, error)
	HelperUpdate(helperType, helperID string, params map[string]interface{}) (map[string]interface{}, error)
	HelperDelete(helperType, helperID string) error
	DeleteHelperByEntityOrEntryID(id string, helperType string) error

	// Config flows (WS-based)
	ConfigFlowInit(handler string, context map[string]interface{}) (map[string]interface{}, error)
	ConfigFlowConfigure(flowID string, data map[string]interface{}) (map[string]interface{}, error)

	// Config entries
	ConfigEntriesList(domain string) ([]interface{}, error)
	ConfigEntryDelete(entryID string) error
	ResolveEntityToConfigEntry(entityID string) (string, error)
}

// RestAPI defines the interface for REST operations against Home Assistant.
// This enables unit-testing command handlers with mock implementations.
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

	// Config flow methods
	ConfigFlowCreate(handler string) (map[string]interface{}, error)
	ConfigFlowStep(flowID string, data map[string]interface{}) (map[string]interface{}, error)
	ConfigEntryDelete(entryID string) error
}
