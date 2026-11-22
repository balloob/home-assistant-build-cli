package cmd

import "github.com/spf13/cobra"

func init() {
	flow := managementGroup(integrationCmd, "flow", "Set up integrations and resume discovery or reauthentication")
	addManagementCommands(flow, []managementSpec{
		{Use: "list", Short: "List pending integration flows, including reauthentication", Transport: "ws", Endpoint: "config_entries/flow/progress", Effect: "read"},
	})
	addManagementFlow(flow, "config/config_entries/flow", managementSpec{Use: "start <handler>", Short: "Start integration setup; supply entry_id to reconfigure", Args: []string{"handler"}, Input: true})
	options := managementGroup(integrationCmd, "options", "Run integration options flows")
	addManagementFlow(options, "config/config_entries/options/flow", managementSpec{Use: "start <handler>", Short: "Start options for a config entry ID", Args: []string{"handler"}})
	subentry := managementGroup(integrationSubentryCmd, "flow", "Set up or reconfigure integration subentries")
	addManagementFlow(subentry, "config/config_entries/subentries/flow", managementSpec{Use: "start <entry_id> <subentry_type>", Short: "Start a subentry flow; supply subentry_id to reconfigure", Args: []string{"entry_id", "subentry_type"}, Input: true, Prepare: func(data map[string]any) error {
		data["handler"] = []any{data["entry_id"], data["subentry_type"]}
		delete(data, "entry_id")
		delete(data, "subentry_type")
		return nil
	}})
	repair := managementGroup(repairsCmd, "flow", "Run repair flows")
	addManagementFlow(repair, "repairs/issues/fix", managementSpec{Use: "start <handler> <issue_id>", Short: "Start a repair using its domain and issue ID", Args: []string{"handler", "issue_id"}})
}

func addManagementFlow(parent *cobra.Command, endpoint string, start managementSpec) {
	start.Transport, start.Method, start.Endpoint, start.Effect = "rest", "POST", endpoint, "write"
	addManagementCommands(parent, []managementSpec{
		start,
		{Use: "get <flow_id>", Short: "Read the current flow step", Transport: "rest", Method: "GET", Endpoint: endpoint + "/{flow_id}", Args: []string{"flow_id"}, Effect: "read"},
		{Use: "step <flow_id>", Short: "Submit fields for the current step; use --data '{}' for an empty form", Transport: "rest", Method: "POST", Endpoint: endpoint + "/{flow_id}", Args: []string{"flow_id"}, Input: true, RequireInput: true, Effect: "write"},
		{Use: "abort <flow_id>", Short: "Abort an unfinished flow", Transport: "rest", Method: "DELETE", Endpoint: endpoint + "/{flow_id}", Args: []string{"flow_id"}, Effect: "write"},
	})
}
