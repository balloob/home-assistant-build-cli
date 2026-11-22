package cmd

func init() {
	addManagementCommands(statisticsCmd, []managementSpec{
		{Use: "query", Short: "Query aggregated statistics", Transport: "ws", Endpoint: "recorder/statistics_during_period", Effect: "read", Input: true, RequireInput: true, Required: []string{"start_time", "period"}},
		{Use: "summary <statistic_id>", Short: "Query one statistic over a fixed or relative period", Transport: "ws", Endpoint: "recorder/statistic_during_period", Effect: "read", Args: []string{"statistic_id"}, Input: true, RequireInput: true},
		{Use: "validate", Short: "List statistics issues", Transport: "ws", Endpoint: "recorder/validate_statistics", Effect: "read"},
		{Use: "refresh-issues", Short: "Refresh statistics repair issues", Transport: "ws", Endpoint: "recorder/update_statistics_issues", Effect: "write"},
		{Use: "adjust-sum <statistic_id>", Short: "Adjust a cumulative statistic from a start time", Transport: "ws", Endpoint: "recorder/adjust_sum_statistics", Effect: "destructive", Args: []string{"statistic_id"}, Input: true, RequireInput: true, Required: []string{"start_time", "adjustment", "adjustment_unit_of_measurement"}},
		{Use: "clear", Short: "Delete statistics for selected IDs and an optional period", Transport: "ws", Endpoint: "recorder/clear_statistics", Effect: "destructive", Input: true, RequireInput: true, Required: []string{"statistic_ids"}},
	})
	addManagementCommands(statisticsMetadataCmd, []managementSpec{
		{Use: "list", Short: "Read statistics metadata; optionally filter statistic_ids", Transport: "ws", Endpoint: "recorder/get_statistics_metadata", Effect: "read", Input: true},
		{Use: "update <statistic_id>", Short: "Correct statistics units and unit class", Transport: "ws", Endpoint: "recorder/update_statistics_metadata", Effect: "write", Args: []string{"statistic_id"}, Input: true, RequireInput: true, Required: []string{"unit_of_measurement"}},
	})
}
