package cmd

func init() {
	addManagementCommands(appCmd, []managementSpec{
		{Use: "logs <slug>", Short: "Read recent app logs", Transport: "rest", Method: "GET", Endpoint: "hassio/addons/{slug}/logs", Effect: "read", Args: []string{"slug"}},
		{Use: "stats <slug>", Short: "Read app resource usage", Transport: "supervisor", Method: "get", Endpoint: "/addons/{slug}/stats", Effect: "read", Args: []string{"slug"}},
		{Use: "validate <slug>", Short: "Validate an app options object", Transport: "supervisor", Method: "post", Endpoint: "/addons/{slug}/options/validate", Effect: "read", Args: []string{"slug"}, Input: true, RequireInput: true},
		{Use: "security <slug>", Short: "Set app protected mode", Transport: "supervisor", Method: "post", Endpoint: "/addons/{slug}/security", Effect: "destructive", Args: []string{"slug"}, Input: true, RequireInput: true, Required: []string{"protected"}},
		{Use: "rebuild <slug>", Short: "Rebuild an app", Transport: "supervisor", Method: "post", Endpoint: "/addons/{slug}/rebuild", Effect: "destructive", Args: []string{"slug"}, Input: true},
		{Use: "refresh", Short: "Refresh the app store", Transport: "supervisor", Method: "post", Endpoint: "/store/reload", Effect: "write"},
	})
	addManagementCommands(appRepositoryCmd, []managementSpec{
		{Use: "list", Short: "List app repositories", Transport: "supervisor", Method: "get", Endpoint: "/store/repositories", Effect: "read"},
		{Use: "add <repository>", Short: "Add an app repository URL", Transport: "supervisor", Method: "post", Endpoint: "/store/repositories", Effect: "write", Args: []string{"repository"}},
		{Use: "remove <repository>", Short: "Remove an app repository by slug", Transport: "supervisor", Method: "delete", Endpoint: "/store/repositories/{repository}", Effect: "destructive", Args: []string{"repository"}},
	})
}
