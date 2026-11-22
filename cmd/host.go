package cmd

func init() {
	addManagementCommands(hostCmd, []managementSpec{
		{Use: "info", Short: "Read host information", Transport: "supervisor", Method: "get", Endpoint: "/host/info", Effect: "read"},
		{Use: "reboot", Short: "Reboot the host", Transport: "supervisor", Method: "post", Endpoint: "/host/reboot", Effect: "destructive"},
		{Use: "shutdown", Short: "Shut down the host", Transport: "supervisor", Method: "post", Endpoint: "/host/shutdown", Effect: "destructive"},
		{Use: "options", Short: "Set host options", Transport: "supervisor", Method: "post", Endpoint: "/host/options", Effect: "write", Input: true, RequireInput: true},
	})
	addManagementCommands(hostOSCmd, []managementSpec{
		{Use: "info", Short: "Read operating system information", Transport: "supervisor", Method: "get", Endpoint: "/os/info", Effect: "read"},
		{Use: "update", Short: "Update Home Assistant OS", Transport: "supervisor", Method: "post", Endpoint: "/os/update", Effect: "destructive", Input: true},
		{Use: "sync", Short: "Import OS configuration from external media", Transport: "supervisor", Method: "post", Endpoint: "/os/config/sync", Effect: "destructive"},
	})
	addManagementCommands(hostDisksCmd, []managementSpec{
		{Use: "list", Short: "List host disks", Transport: "supervisor", Method: "get", Endpoint: "/host/disks", Effect: "read"},
		{Use: "usage <disk>", Short: "Read disk usage", Transport: "supervisor", Method: "get", Endpoint: "/host/disks/{disk}/usage", Effect: "read", Args: []string{"disk"}},
	})
	addManagementCommands(hostDatadiskCmd, []managementSpec{
		{Use: "list", Short: "List eligible data disks", Transport: "supervisor", Method: "get", Endpoint: "/os/datadisk/list", Effect: "read"},
		{Use: "move <device>", Short: "Move the data partition to a device", Transport: "supervisor", Method: "post", Endpoint: "/os/datadisk/move", Effect: "destructive", Args: []string{"device"}},
	})
	addManagementCommands(networkInterfacesCmd, []managementSpec{
		{Use: "list", Short: "Read host IP and Wi-Fi settings", Transport: "supervisor", Method: "get", Endpoint: "/network/info", Effect: "read"},
		{Use: "update <interface>", Short: "Change host IP or Wi-Fi settings", Transport: "supervisor", Method: "post", Endpoint: "/network/interface/{interface}/update", Effect: "destructive", Args: []string{"interface"}, Input: true, RequireInput: true},
		{Use: "scan <interface>", Short: "Scan Wi-Fi access points", Transport: "supervisor", Method: "get", Endpoint: "/network/interface/{interface}/accesspoints", Effect: "read", Args: []string{"interface"}},
	})
	addManagementCommands(mountCmd, []managementSpec{
		{Use: "list", Short: "List storage mounts", Transport: "supervisor", Method: "get", Endpoint: "/mounts", Effect: "read"},
		{Use: "create", Short: "Create a storage mount", Transport: "supervisor", Method: "post", Endpoint: "/mounts", Effect: "write", Input: true, RequireInput: true, Required: []string{"name", "type", "usage"}},
		{Use: "update <name>", Short: "Update a storage mount", Transport: "supervisor", Method: "put", Endpoint: "/mounts/{name}", Effect: "write", Args: []string{"name"}, Input: true, RequireInput: true},
		{Use: "delete <name>", Short: "Delete a storage mount", Transport: "supervisor", Method: "delete", Endpoint: "/mounts/{name}", Effect: "destructive", Args: []string{"name"}},
		{Use: "reload <name>", Short: "Reload a storage mount", Transport: "supervisor", Method: "post", Endpoint: "/mounts/{name}/reload", Effect: "write", Args: []string{"name"}},
		{Use: "options", Short: "Set storage mount defaults", Transport: "supervisor", Method: "post", Endpoint: "/mounts/options", Effect: "write", Input: true, RequireInput: true},
	})
}
