package client

import (
	"fmt"
	"time"
)

// High-level API methods

// GetStates returns all entity states
func (c *WebSocketClient) GetStates() ([]interface{}, error) {
	return c.sendListCommand("get_states", nil)
}

// GetConfig returns the Home Assistant configuration
func (c *WebSocketClient) GetConfig() (map[string]interface{}, error) {
	return c.sendMapCommand("get_config", nil)
}

// GetServices returns all available services
func (c *WebSocketClient) GetServices() (map[string]interface{}, error) {
	return c.sendMapCommand("get_services", nil)
}

// CallService calls a service
func (c *WebSocketClient) CallService(domain, service string, data, target map[string]interface{}, returnResponse bool) (interface{}, error) {
	params := map[string]interface{}{
		"domain":  domain,
		"service": service,
	}
	if data != nil {
		params["service_data"] = data
	}
	if target != nil {
		params["target"] = target
	}
	if returnResponse {
		params["return_response"] = true
	}
	return c.SendCommand("call_service", params)
}

// Ping sends a ping message
func (c *WebSocketClient) Ping() error {
	_, err := c.SendCommand("ping", nil)
	return err
}

// Registry operations

// AreaRegistryList returns all areas
func (c *WebSocketClient) AreaRegistryList() ([]interface{}, error) {
	return c.sendListCommand("config/area_registry/list", nil)
}

// AreaRegistryCreate creates a new area
func (c *WebSocketClient) AreaRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/area_registry/create", "name", name, params)
}

// AreaRegistryUpdate updates an area
func (c *WebSocketClient) AreaRegistryUpdate(areaID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/area_registry/update", "area_id", areaID, params)
}

// AreaRegistryDelete deletes an area
func (c *WebSocketClient) AreaRegistryDelete(areaID string) error {
	return c.sendDelete("config/area_registry/delete", "area_id", areaID)
}

// FloorRegistryList returns all floors
func (c *WebSocketClient) FloorRegistryList() ([]interface{}, error) {
	return c.sendListCommand("config/floor_registry/list", nil)
}

// FloorRegistryCreate creates a new floor
func (c *WebSocketClient) FloorRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/floor_registry/create", "name", name, params)
}

// FloorRegistryUpdate updates a floor
func (c *WebSocketClient) FloorRegistryUpdate(floorID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/floor_registry/update", "floor_id", floorID, params)
}

// FloorRegistryDelete deletes a floor
func (c *WebSocketClient) FloorRegistryDelete(floorID string) error {
	return c.sendDelete("config/floor_registry/delete", "floor_id", floorID)
}

// LabelRegistryList returns all labels
func (c *WebSocketClient) LabelRegistryList() ([]interface{}, error) {
	return c.sendListCommand("config/label_registry/list", nil)
}

// LabelRegistryCreate creates a new label
func (c *WebSocketClient) LabelRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/label_registry/create", "name", name, params)
}

// LabelRegistryUpdate updates a label
func (c *WebSocketClient) LabelRegistryUpdate(labelID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/label_registry/update", "label_id", labelID, params)
}

// LabelRegistryDelete deletes a label
func (c *WebSocketClient) LabelRegistryDelete(labelID string) error {
	return c.sendDelete("config/label_registry/delete", "label_id", labelID)
}

// DeviceRegistryList returns all devices
func (c *WebSocketClient) DeviceRegistryList() ([]interface{}, error) {
	return c.sendListCommand("config/device_registry/list", nil)
}

// DeviceRegistryUpdate updates a device
func (c *WebSocketClient) DeviceRegistryUpdate(deviceID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/device_registry/update", "device_id", deviceID, params)
}

// EntityRegistryList returns all entities
func (c *WebSocketClient) EntityRegistryList() ([]interface{}, error) {
	return c.sendListCommand("config/entity_registry/list", nil)
}

// EntityRegistryGet returns a specific entity
func (c *WebSocketClient) EntityRegistryGet(entityID string) (map[string]interface{}, error) {
	return c.sendMapCommand("config/entity_registry/get", map[string]interface{}{
		"entity_id": entityID,
	})
}

// EntityRegistryUpdate updates an entity
func (c *WebSocketClient) EntityRegistryUpdate(entityID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/entity_registry/update", "entity_id", entityID, params)
}

// ZoneList returns all zones
func (c *WebSocketClient) ZoneList() ([]interface{}, error) {
	return c.sendListCommand("zone/list", nil)
}

// ZoneCreate creates a new zone
func (c *WebSocketClient) ZoneCreate(name string, latitude, longitude, radius float64, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{
		"name":      name,
		"latitude":  latitude,
		"longitude": longitude,
		"radius":    radius,
	}
	for k, v := range params {
		p[k] = v
	}
	return c.sendMapCommand("zone/create", p)
}

// ZoneUpdate updates a zone
func (c *WebSocketClient) ZoneUpdate(zoneID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("zone/update", "zone_id", zoneID, params)
}

// ZoneDelete deletes a zone
func (c *WebSocketClient) ZoneDelete(zoneID string) error {
	return c.sendDelete("zone/delete", "zone_id", zoneID)
}

// SystemHealthInfo returns system health information using subscription
func (c *WebSocketClient) SystemHealthInfo() (map[string]interface{}, error) {
	if err := c.requireAuth(); err != nil {
		return nil, err
	}

	// Channel to receive events
	eventCh := make(chan map[string]interface{}, 100)
	doneCh := make(chan struct{})

	// Accumulated data
	data := make(map[string]interface{})
	var dataErr error

	// Assign ID, register subscription, and send — all under the write
	// lock to guarantee monotonically increasing IDs on the wire.
	c.writeMu.Lock()
	msgID := c.nextID()

	c.subsMu.Lock()
	c.subscriptions[msgID] = func(event map[string]interface{}) {
		select {
		case eventCh <- event:
		default:
		}
	}
	c.subsMu.Unlock()

	msg := map[string]interface{}{
		"id":   msgID,
		"type": "system_health/info",
	}

	respCh := make(chan *WSMessage, 1)
	c.pendingMu.Lock()
	c.pending[msgID] = respCh
	c.pendingMu.Unlock()

	writeErr := c.conn.WriteJSON(msg)
	c.writeMu.Unlock()

	// Cleanup subscription on exit
	defer func() {
		c.subsMu.Lock()
		delete(c.subscriptions, msgID)
		c.subsMu.Unlock()
	}()

	if writeErr != nil {
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to send command: %s", writeErr)}
	}

	// Wait for initial result (subscription confirmation)
	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: "connection closed"}
		}
		if !resp.Success {
			return nil, wsResponseError(resp)
		}
	case <-time.After(c.Timeout):
		return nil, &APIError{Code: ErrCodeTimeout, Message: "timeout waiting for subscription confirmation"}
	}

	// Process events in a goroutine
	go aggregateHealthEvents(eventCh, data, doneCh)

	// Wait for finish or timeout
	select {
	case <-doneCh:
		// Normal completion
	case <-time.After(c.Timeout):
		dataErr = &APIError{Code: ErrCodeTimeout, Message: "timeout waiting for system health data"}
	}

	close(eventCh)

	if dataErr != nil {
		return nil, dataErr
	}

	return data, nil
}

// aggregateHealthEvents processes system_health/info subscription events,
// merging them into the provided data map. It reads from eventCh and closes
// doneCh when a "finish" event is received (or the channel is closed).
func aggregateHealthEvents(eventCh <-chan map[string]interface{}, data map[string]interface{}, doneCh chan struct{}) {
	defer close(doneCh)
	for event := range eventCh {
		eventType, _ := event["type"].(string)

		switch eventType {
		case "initial":
			if eventData, ok := event["data"].(map[string]interface{}); ok {
				for k, v := range eventData {
					data[k] = v
				}
			}
		case "update":
			domain, _ := event["domain"].(string)
			key, _ := event["key"].(string)
			success, _ := event["success"].(bool)

			if domain != "" && key != "" {
				if _, exists := data[domain]; !exists {
					data[domain] = map[string]interface{}{
						"info": make(map[string]interface{}),
					}
				}
				if domainData, ok := data[domain].(map[string]interface{}); ok {
					if _, exists := domainData["info"]; !exists {
						domainData["info"] = make(map[string]interface{})
					}
					if infoData, ok := domainData["info"].(map[string]interface{}); ok {
						if success {
							infoData[key] = event["data"]
						} else {
							if errData, ok := event["error"].(map[string]interface{}); ok {
								infoData[key] = map[string]interface{}{
									"error": true,
									"value": errData["msg"],
								}
							}
						}
					}
				}
			}
		case "finish":
			return
		}
	}
}

// SearchRelated returns related items for a given item type and ID
// itemType can be: area, automation, automation_blueprint, config_entry, device, entity, floor, group, label, scene, script, script_blueprint
func (c *WebSocketClient) SearchRelated(itemType, itemID string) (map[string][]string, error) {
	result, err := c.SendCommand("search/related", map[string]interface{}{
		"item_type": itemType,
		"item_id":   itemID,
	})
	if err != nil {
		return nil, err
	}

	// Convert the result to a map of string slices
	resultMap := make(map[string][]string)
	if m, ok := result.(map[string]interface{}); ok {
		for key, value := range m {
			if arr, ok := value.([]interface{}); ok {
				items := make([]string, 0, len(arr))
				for _, item := range arr {
					if str, ok := item.(string); ok {
						items = append(items, str)
					}
				}
				resultMap[key] = items
			}
		}
	}

	return resultMap, nil
}

// Helper operations

// HelperList returns all helpers of a specific type
// helperType can be: input_boolean, input_number, input_text, input_select, input_datetime, input_button, counter, timer, schedule
func (c *WebSocketClient) HelperList(helperType string) ([]interface{}, error) {
	return c.sendListCommand(helperType+"/list", nil)
}

// HelperCreate creates a new helper of a specific type
// helperType can be: input_boolean, input_number, input_text, input_select, input_datetime, input_button, counter, timer, schedule
func (c *WebSocketClient) HelperCreate(helperType string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.sendMapCommand(helperType+"/create", params)
}

// HelperUpdate updates an existing helper
// helperType can be: input_boolean, input_number, input_text, input_select, input_datetime, input_button, counter, timer, schedule
// The idField depends on the helper type (e.g., "input_boolean_id", "counter_id", "timer_id", "schedule_id")
func (c *WebSocketClient) HelperUpdate(helperType, helperID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend(helperType+"/update", helperType+"_id", helperID, params)
}

// HelperDelete deletes a helper
// helperType can be: input_boolean, input_number, input_text, input_select, input_datetime, input_button, counter, timer, schedule
// The idField depends on the helper type (e.g., "input_boolean_id", "counter_id", "timer_id", "schedule_id")
func (c *WebSocketClient) HelperDelete(helperType, helperID string) error {
	return c.sendDelete(helperType+"/delete", helperType+"_id", helperID)
}

// Config Entry Flow operations

// ConfigFlowInit starts a new config flow for an integration
func (c *WebSocketClient) ConfigFlowInit(handler string, context map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"handler": handler,
	}
	if context != nil {
		params["context"] = context
	}
	return c.sendMapCommand("config_entries/flow", params)
}

// ConfigFlowConfigure submits data to a config flow step.
// The flow_id is set after merging data to prevent accidental overwrite.
func (c *WebSocketClient) ConfigFlowConfigure(flowID string, data map[string]interface{}) (map[string]interface{}, error) {
	params := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		params[k] = v
	}
	params["flow_id"] = flowID // set last to prevent collision
	return c.sendMapCommand("config_entries/flow", params)
}

// ConfigEntriesList returns all config entries, optionally filtered by domain
func (c *WebSocketClient) ConfigEntriesList(domain string) ([]interface{}, error) {
	params := map[string]interface{}{}
	if domain != "" {
		params["domain"] = domain
	}
	return c.sendListCommand("config_entries/get", params)
}

// ConfigEntryDelete deletes a config entry
func (c *WebSocketClient) ConfigEntryDelete(entryID string) error {
	_, err := c.SendCommand("config_entries/delete", map[string]interface{}{
		"entry_id": entryID,
	})
	return err
}

// ResolveEntityToConfigEntry resolves an entity_id to its config_entry_id
// Returns the config_entry_id if found, or empty string if the entity doesn't have one
func (c *WebSocketClient) ResolveEntityToConfigEntry(entityID string) (string, error) {
	entity, err := c.EntityRegistryGet(entityID)
	if err != nil {
		return "", err
	}
	if configEntryID, ok := entity["config_entry_id"].(string); ok && configEntryID != "" {
		return configEntryID, nil
	}
	return "", nil
}

// Person registry operations

// PersonRegistryList returns all person entries
func (c *WebSocketClient) PersonRegistryList() ([]interface{}, error) {
	result, err := c.SendCommand("person/list", nil)
	if err != nil {
		return nil, err
	}
	// HA returns {"storage": [...], "config": [...]} — merge both lists
	if m, ok := result.(map[string]interface{}); ok {
		var persons []interface{}
		if storage, ok := m["storage"].([]interface{}); ok {
			persons = append(persons, storage...)
		}
		if config, ok := m["config"].([]interface{}); ok {
			persons = append(persons, config...)
		}
		return persons, nil
	}
	if list, ok := result.([]interface{}); ok {
		return list, nil
	}
	return nil, fmt.Errorf("unexpected response from person/list")
}

// PersonRegistryCreate creates a new person
func (c *WebSocketClient) PersonRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("person/create", "name", name, params)
}

// PersonRegistryUpdate updates a person
func (c *WebSocketClient) PersonRegistryUpdate(personID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("person/update", "id", personID, params)
}

// PersonRegistryDelete deletes a person
func (c *WebSocketClient) PersonRegistryDelete(personID string) error {
	return c.sendDelete("person/delete", "id", personID)
}

// Category registry operations

// CategoryRegistryList returns all categories for a given scope
func (c *WebSocketClient) CategoryRegistryList(scope string) ([]interface{}, error) {
	params := map[string]interface{}{"scope": scope}
	return c.sendListCommand("config/category_registry/list", params)
}

// CategoryRegistryCreate creates a new category in a scope
func (c *WebSocketClient) CategoryRegistryCreate(scope, name string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{
		"scope": scope,
		"name":  name,
	}
	for k, v := range params {
		p[k] = v
	}
	return c.sendMapCommand("config/category_registry/create", p)
}

// CategoryRegistryUpdate updates a category
func (c *WebSocketClient) CategoryRegistryUpdate(categoryID string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.mergeAndSend("config/category_registry/update", "category_id", categoryID, params)
}

// CategoryRegistryDelete deletes a category
func (c *WebSocketClient) CategoryRegistryDelete(scope, categoryID string) error {
	_, err := c.SendCommand("config/category_registry/delete", map[string]interface{}{
		"scope":       scope,
		"category_id": categoryID,
	})
	return err
}

// Integration (config entry) operations

// ConfigEntryGet returns a single config entry by entry_id
func (c *WebSocketClient) ConfigEntryGet(entryID string) (map[string]interface{}, error) {
	result, err := c.SendCommand("config_entries/get_single", map[string]interface{}{
		"entry_id": entryID,
	})
	if err != nil {
		return nil, err
	}
	// Response is {"config_entry": {...}}
	if m, ok := result.(map[string]interface{}); ok {
		if entry, ok := m["config_entry"].(map[string]interface{}); ok {
			return entry, nil
		}
		return nil, fmt.Errorf("unexpected response from config_entries/get_single: missing config_entry key")
	}
	return nil, fmt.Errorf("unexpected response from config_entries/get_single")
}

// ConfigEntrySetDisabled enables or disables a config entry.
// Pass "user" to disable, nil to enable.
func (c *WebSocketClient) ConfigEntrySetDisabled(entryID string, disabledBy interface{}) (map[string]interface{}, error) {
	return c.sendMapCommand("config_entries/disable", map[string]interface{}{
		"entry_id":    entryID,
		"disabled_by": disabledBy,
	})
}

// Todo item operations

// TodoItemList returns the items for a to-do list entity
func (c *WebSocketClient) TodoItemList(entityID string) ([]interface{}, error) {
	result, err := c.SendCommand("todo/item/list", map[string]interface{}{
		"entity_id": entityID,
	})
	if err != nil {
		return nil, err
	}
	// Response is {"items": [...]}
	if m, ok := result.(map[string]interface{}); ok {
		if items, ok := m["items"].([]interface{}); ok {
			return items, nil
		}
		return []interface{}{}, nil
	}
	return nil, fmt.Errorf("unexpected response from todo/item/list")
}

// Persistent notification operations

// NotificationList returns all current persistent notifications
func (c *WebSocketClient) NotificationList() ([]interface{}, error) {
	result, err := c.SendCommand("persistent_notification/get", nil)
	if err != nil {
		return nil, err
	}
	// Response is a list of notification objects
	if list, ok := result.([]interface{}); ok {
		return list, nil
	}
	// May also come back as a map keyed by notification_id
	if m, ok := result.(map[string]interface{}); ok {
		items := make([]interface{}, 0, len(m))
		for _, v := range m {
			items = append(items, v)
		}
		return items, nil
	}
	return []interface{}{}, nil
}

// Repairs operations

// RepairListIssues returns all active repair issues
func (c *WebSocketClient) RepairListIssues() ([]interface{}, error) {
	result, err := c.SendCommand("repairs/list_issues", nil)
	if err != nil {
		return nil, err
	}
	// Response is {"issues": [...]}
	if m, ok := result.(map[string]interface{}); ok {
		if issues, ok := m["issues"].([]interface{}); ok {
			return issues, nil
		}
		return []interface{}{}, nil
	}
	return nil, fmt.Errorf("unexpected response from repairs/list_issues")
}

// RepairIgnoreIssue ignores or un-ignores a repair issue
func (c *WebSocketClient) RepairIgnoreIssue(domain, issueID string, ignore bool) error {
	_, err := c.SendCommand("repairs/ignore_issue", map[string]interface{}{
		"domain":   domain,
		"issue_id": issueID,
		"ignore":   ignore,
	})
	return err
}

