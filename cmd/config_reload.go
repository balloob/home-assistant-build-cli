package cmd

import (
	"time"

	"github.com/home-assistant/hab/client"
	log "github.com/sirupsen/logrus"
)

// configReloadEvents maps a config API prefix to the event that Home Assistant
// fires when the reload after a save is done.
var configReloadEvents = map[string]string{
	"config/automation/config/": "automation_reloaded",
	"config/scene/config/":      "scene_reloaded",
}

// configReloadTimeout bounds the wait for the reload after a save.
var configReloadTimeout = 30 * time.Second

type configPoster interface {
	Post(endpoint string, body interface{}) (interface{}, error)
	CallService(domain, service string, data map[string]interface{}) (interface{}, error)
}

// postConfig saves a config through the config API and returns when Home
// Assistant has reloaded it. The API answers before the reload is done,
// because it starts the reload as a background task.
func postConfig(rest configPoster, prefix, id string, config interface{}) (interface{}, error) {
	if _, ok := configReloadEvents[prefix]; !ok {
		return postConfigAndWait(rest, nil, prefix, id, config)
	}
	ws, err := getWSClient()
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return postConfigAndWait(rest, nil, prefix, id, config)
	}
	defer ws.Close()
	return postConfigAndWait(rest, ws, prefix, id, config)
}

func postConfigAndWait(rest configPoster, subscriber client.EventSubscriber, prefix, id string, config interface{}) (interface{}, error) {
	if prefix == "config/script/config/" {
		result, err := rest.Post(prefix+id, config)
		if err != nil {
			return nil, err
		}
		// A script reload fires no event. The REST service call returns when
		// the reload is done.
		if _, err := rest.CallService("script", "reload", nil); err != nil {
			return nil, err
		}
		return result, nil
	}

	eventType, ok := configReloadEvents[prefix]
	if !ok || subscriber == nil {
		return rest.Post(prefix+id, config)
	}

	events, unsubscribe, err := subscriber.SubscribeEvents(eventType)
	if err != nil {
		return nil, err
	}
	defer unsubscribe()

	result, err := rest.Post(prefix+id, config)
	if err != nil {
		return nil, err
	}

	select {
	case <-events:
	case <-time.After(configReloadTimeout):
		log.Warnf("Home Assistant did not report %s within %s", eventType, configReloadTimeout)
	}
	return result, nil
}
