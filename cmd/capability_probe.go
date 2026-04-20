package cmd

import (
	"errors"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

type capabilityStatus struct {
	Available  bool   `json:"available"`
	Restricted bool   `json:"restricted,omitempty"`
	Reason     string `json:"reason,omitempty"`
	Error      string `json:"error,omitempty"`
}

var capabilityProbeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Probe environment capabilities",
	Long:  "Probe authenticated runtime capabilities (REST, WebSocket, supervisor, ESPHome) and derived workflow support.",
	RunE:  runCapabilityProbe,
}

func init() {
	capabilityCmd.AddCommand(capabilityProbeCmd)
	mergeSchemaAnnotation(capabilityProbeCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"local"},
		ResourceType: "capability",
		GuideTopic:   "discovery",
	})
}

func runCapabilityProbe(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()
	manager := getAuthManager()

	authStatus := manager.GetAuthStatus()
	authenticated, _ := authStatus["authenticated"].(bool)

	restStatus := capabilityStatus{Available: false, Reason: "authentication required"}
	wsStatus := capabilityStatus{Available: false, Reason: "authentication required"}
	supervisorStatus := capabilityStatus{Available: false, Reason: "authentication required"}
	esphomeStatus := capabilityStatus{Available: false, Reason: "authentication required"}
	backupStatus := capabilityStatus{Available: false, Reason: "websocket unavailable"}
	diagnosticsStatus := capabilityStatus{Available: false, Reason: "websocket unavailable"}
	networkStatus := capabilityStatus{Available: false, Reason: "websocket unavailable"}
	energyStatus := capabilityStatus{Available: false, Reason: "websocket unavailable"}
	repairsStatus := capabilityStatus{Available: false, Reason: "websocket unavailable"}

	environment := map[string]any{
		"supervisor_env": auth.IsSupervisorEnvironment(),
	}

	if authenticated {
		restClient, err := getRESTClient()
		if err != nil {
			restStatus = statusFromError(err)
		} else {
			config, cfgErr := restClient.GetConfig()
			if cfgErr != nil {
				restStatus = statusFromError(cfgErr)
			} else {
				restStatus = capabilityStatus{Available: true}
				environment["ha_version"] = config["version"]
				environment["ha_state"] = config["state"]
				environment["location_name"] = config["location_name"]
			}

			supervisorStatus = probeSupervisor(restClient)
		}

		ws, wsErr := getWSClient()
		if wsErr != nil {
			wsStatus = statusFromError(wsErr)
		} else {
			defer ws.Close()
			wsStatus = probeStatus(func() error {
				_, err := ws.GetConfig()
				return err
			})
			backupStatus = probeStatus(func() error {
				_, err := ws.BackupInfo()
				return err
			})
			diagnosticsStatus = probeStatus(func() error {
				_, err := ws.DiagnosticsList()
				return err
			})
			networkStatus = probeStatus(func() error {
				_, err := ws.NetworkGet()
				return err
			})
			energyStatus = probeStatus(func() error {
				_, err := ws.EnergyInfo()
				return err
			})
			repairsStatus = probeStatus(func() error {
				_, err := ws.RepairListIssues()
				return err
			})
		}

		esClient, esErr := getESPHomeClient()
		if esErr != nil {
			esphomeStatus = statusFromError(esErr)
		} else {
			version, verErr := esClient.GetVersion()
			if verErr != nil {
				esphomeStatus = statusFromError(verErr)
			} else {
				esphomeStatus = capabilityStatus{Available: true}
				environment["esphome_version"] = version
			}
		}
	}

	if !supervisorStatus.Available && auth.IsSupervisorEnvironment() {
		supervisorStatus = capabilityStatus{Available: true}
	}
	if !supervisorStatus.Available && anyCapabilityAvailable(backupStatus, diagnosticsStatus, networkStatus, repairsStatus) {
		supervisorStatus = capabilityStatus{Available: true}
	}

	result := map[string]any{
		"auth":        authStatus,
		"environment": environment,
		"checks": map[string]any{
			"rest":        restStatus,
			"websocket":   wsStatus,
			"supervisor":  supervisorStatus,
			"esphome":     esphomeStatus,
			"backup":      backupStatus,
			"diagnostics": diagnosticsStatus,
			"network":     networkStatus,
			"energy":      energyStatus,
			"repairs":     repairsStatus,
		},
		"capabilities": map[string]any{
			"auth":                                authenticated,
			"rest":                                restStatus.Available,
			"ws":                                  wsStatus.Available,
			"supervisor":                          supervisorStatus.Available,
			"esphome":                             esphomeStatus.Available,
			"can_manage_backups":                  backupStatus.Available,
			"can_use_diagnostics":                 diagnosticsStatus.Available,
			"can_manage_network":                  networkStatus.Available,
			"can_manage_repairs":                  repairsStatus.Available,
			"supports_energy_preferences":         energyStatus.Available,
			"requires_supervisor_for_backup":      true,
			"requires_supervisor_for_network":     true,
			"requires_supervisor_for_repairs":     true,
			"requires_supervisor_for_diagnostics": true,
		},
	}

	output.PrintOutputWithContext(result, textMode, "", output.EnvelopeContext{
		Operation:    "probe",
		ResourceType: "capability",
	})
	return nil
}

func probeSupervisor(restClient client.RestAPI) capabilityStatus {
	if restClient == nil {
		return capabilityStatus{Available: false, Reason: "rest unavailable"}
	}
	_, err := restClient.Get("hassio/info")
	if err == nil {
		return capabilityStatus{Available: true}
	}
	status := statusFromError(err)
	if isUnsupportedError(err) {
		status.Reason = "supervisor endpoint unavailable"
		status.Error = ""
	}
	return status
}

func probeStatus(check func() error) capabilityStatus {
	err := check()
	if err == nil {
		return capabilityStatus{Available: true}
	}
	return statusFromError(err)
}

func anyCapabilityAvailable(statuses ...capabilityStatus) bool {
	for _, status := range statuses {
		if status.Available {
			return true
		}
	}
	return false
}

func statusFromError(err error) capabilityStatus {
	if err == nil {
		return capabilityStatus{Available: true}
	}
	if isUnsupportedError(err) {
		return capabilityStatus{Available: false, Reason: "unsupported"}
	}
	if isRestrictedError(err) {
		return capabilityStatus{Available: true, Restricted: true, Reason: "permission restricted"}
	}
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		return capabilityStatus{Available: false, Reason: apiErr.Code, Error: apiErr.Message}
	}
	return capabilityStatus{Available: false, Error: err.Error()}
}

func isRestrictedError(err error) bool {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == client.ErrCodePermissionDenied
	}
	return false
}

func isUnsupportedError(err error) bool {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code == client.ErrCodeNotFound {
			return true
		}
		msg := strings.ToLower(apiErr.Message)
		if strings.Contains(msg, "unknown command") || strings.Contains(msg, "not supported") || strings.Contains(msg, "not found") {
			return true
		}
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown command") || strings.Contains(msg, "not supported") || strings.Contains(msg, "404")
}
