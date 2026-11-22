package cmd

func init() {
	addManagementCommands(userCmd, []managementSpec{
		{Use: "list", Short: "List Home Assistant users", Transport: "ws", Endpoint: "config/auth/list", Effect: "read"},
		{Use: "create", Short: "Create a user; add login credentials separately", Transport: "ws", Endpoint: "config/auth/create", Effect: "write", Input: true, RequireInput: true, Required: []string{"name"}},
		{Use: "update <user_id>", Short: "Update a user profile and permissions", Transport: "ws", Endpoint: "config/auth/update", Effect: "write", Args: []string{"user_id"}, Input: true, RequireInput: true},
		{Use: "delete <user_id>", Short: "Delete a user", Transport: "ws", Endpoint: "config/auth/delete", Effect: "destructive", Args: []string{"user_id"}},
		{Use: "password <user_id>", Short: "Set a user password as an administrator", Transport: "ws", Endpoint: "config/auth_provider/homeassistant/admin_change_password", Effect: "destructive", Args: []string{"user_id"}, Input: true, RequireInput: true, Required: []string{"password"}},
		{Use: "username <user_id>", Short: "Change a login username", Transport: "ws", Endpoint: "config/auth_provider/homeassistant/admin_change_username", Effect: "write", Args: []string{"user_id"}, Input: true, RequireInput: true, Required: []string{"username"}},
	})
	addManagementCommands(userCredentialsCmd, []managementSpec{
		{Use: "create <user_id>", Short: "Add a local username and password to a user", Transport: "ws", Endpoint: "config/auth_provider/homeassistant/create", Effect: "write", Args: []string{"user_id"}, Input: true, RequireInput: true, Required: []string{"username", "password"}},
	})
	addManagementCommands(applicationCredentialCmd, []managementSpec{
		{Use: "list", Short: "List application credentials, including client secrets", Transport: "ws", Endpoint: "application_credentials/list", Effect: "read"},
		{Use: "create", Short: "Create application credentials", Transport: "ws", Endpoint: "application_credentials/create", Effect: "write", Input: true, RequireInput: true, Required: []string{"domain", "client_id", "client_secret"}},
		{Use: "delete <application_credentials_id>", Short: "Delete application credentials", Transport: "ws", Endpoint: "application_credentials/delete", Effect: "destructive", Args: []string{"application_credentials_id"}},
		{Use: "config", Short: "List supported application credential domains", Transport: "ws", Endpoint: "application_credentials/config", Effect: "read"},
		{Use: "entry <config_entry_id>", Short: "Find credentials used by an integration", Transport: "ws", Endpoint: "application_credentials/config_entry", Effect: "read", Args: []string{"config_entry_id"}},
	})
	addManagementCommands(systemConfigCmd, []managementSpec{
		{Use: "get", Short: "Read instance configuration", Transport: "ws", Endpoint: "get_config", Effect: "read"},
		{Use: "update", Short: "Set location, units, time zone, URLs, and locale", Transport: "ws", Endpoint: "config/core/update", Effect: "write", Input: true, RequireInput: true},
		{Use: "detect", Short: "Detect location and unit defaults", Transport: "ws", Endpoint: "config/core/detect", Effect: "read"},
	})
}
