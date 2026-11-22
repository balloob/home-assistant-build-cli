package cmd

func init() {
	addManagementCommands(assistPipelineCmd, []managementSpec{
		{Use: "list", Short: "List Assist pipelines and the preferred pipeline", Transport: "ws", Endpoint: "assist_pipeline/pipeline/list", Effect: "read"},
		{Use: "get [pipeline_id]", Short: "Read an Assist pipeline", Transport: "ws", Endpoint: "assist_pipeline/pipeline/get", Effect: "read", Args: []string{"pipeline_id"}, OptionalArg: true},
		{Use: "create", Short: "Create an Assist pipeline", Transport: "ws", Endpoint: "assist_pipeline/pipeline/create", Effect: "write", Input: true, RequireInput: true, Required: []string{"name", "language", "conversation_engine", "conversation_language", "stt_engine", "stt_language", "tts_engine", "tts_language", "tts_voice", "wake_word_entity", "wake_word_id"}},
		{Use: "update <pipeline_id>", Short: "Replace Assist pipeline settings; supply all pipeline fields", Transport: "ws", Endpoint: "assist_pipeline/pipeline/update", Effect: "write", Args: []string{"pipeline_id"}, Input: true, RequireInput: true, Required: []string{"name", "language", "conversation_engine", "conversation_language", "stt_engine", "stt_language", "tts_engine", "tts_language", "tts_voice", "wake_word_entity", "wake_word_id"}},
		{Use: "delete <pipeline_id>", Short: "Delete an Assist pipeline", Transport: "ws", Endpoint: "assist_pipeline/pipeline/delete", Effect: "destructive", Args: []string{"pipeline_id"}},
		{Use: "set-preferred <pipeline_id>", Short: "Select the preferred Assist pipeline", Transport: "ws", Endpoint: "assist_pipeline/pipeline/set_preferred", Effect: "write", Args: []string{"pipeline_id"}},
	})
	addManagementCommands(assistRunsCmd, []managementSpec{
		{Use: "list <pipeline_id>", Short: "List recent pipeline runs", Transport: "ws", Endpoint: "assist_pipeline/pipeline_debug/list", Effect: "read", Args: []string{"pipeline_id"}},
		{Use: "get <pipeline_id> <pipeline_run_id>", Short: "Read pipeline run events", Transport: "ws", Endpoint: "assist_pipeline/pipeline_debug/get", Effect: "read", Args: []string{"pipeline_id", "pipeline_run_id"}},
	})
	addManagementCommands(assistCmd, []managementSpec{
		{Use: "languages", Short: "List supported Assist languages", Transport: "ws", Endpoint: "assist_pipeline/language/list", Effect: "read"},
		{Use: "devices", Short: "List Assist devices", Transport: "ws", Endpoint: "assist_pipeline/device/list", Effect: "read"},
	})
	addManagementCommands(assistExposureCmd, []managementSpec{
		{Use: "list", Short: "List exposed entities", Transport: "ws", Endpoint: "homeassistant/expose_entity/list", Effect: "read"},
		{Use: "set", Short: "Set entity exposure for Assist, Alexa, or Google Assistant", Transport: "ws", Endpoint: "homeassistant/expose_entity", Effect: "write", Input: true, RequireInput: true, Required: []string{"assistants", "entity_ids", "should_expose"}},
	})
	addManagementCommands(assistExposureDefaultsCmd, []managementSpec{
		{Use: "get <assistant>", Short: "Read exposure defaults", Transport: "ws", Endpoint: "homeassistant/expose_new_entities/get", Effect: "read", Args: []string{"assistant"}},
		{Use: "set <assistant>", Short: "Set exposure defaults", Transport: "ws", Endpoint: "homeassistant/expose_new_entities/set", Effect: "write", Args: []string{"assistant"}, Input: true, RequireInput: true, Required: []string{"expose_new"}},
	})
}
