package cmd

func init() {
	addManagementCommands(integrationCmd, []managementSpec{
		{Use: "types", Short: "List integration flow handlers", Transport: "rest", Method: "GET", Endpoint: "config/config_entries/flow_handlers?type=integration", Effect: "read"},
		{Use: "delete <entry_id>", Short: "Remove an integration", Transport: "rest", Method: "DELETE", Endpoint: "config/config_entries/entry/{entry_id}", Effect: "destructive", Args: []string{"entry_id"}},
		{Use: "update <entry_id>", Short: "Update integration title and preferences", Transport: "ws", Endpoint: "config_entries/update", Effect: "write", Args: []string{"entry_id"}, Input: true, RequireInput: true},
	})
	addManagementCommands(integrationSubentryCmd, []managementSpec{
		{Use: "list <entry_id>", Short: "List integration subentries", Transport: "ws", Endpoint: "config_entries/subentries/list", Effect: "read", Args: []string{"entry_id"}},
		{Use: "update <entry_id> <subentry_id>", Short: "Update a subentry title", Transport: "ws", Endpoint: "config_entries/subentries/update", Effect: "write", Args: []string{"entry_id", "subentry_id"}, Input: true, RequireInput: true, Required: []string{"title"}},
		{Use: "delete <entry_id> <subentry_id>", Short: "Remove an integration subentry", Transport: "ws", Endpoint: "config_entries/subentries/delete", Effect: "destructive", Args: []string{"entry_id", "subentry_id"}},
	})
	addManagementCommands(repairsCmd, []managementSpec{
		{Use: "get <domain> <issue_id>", Short: "Read repair issue data", Transport: "ws", Endpoint: "repairs/get_issue_data", Effect: "read", Args: []string{"domain", "issue_id"}},
	})
}
