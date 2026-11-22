package cmd

func init() {
	addManagementCommands(mediaCmd, []managementSpec{
		{Use: "browse [media_content_id]", Short: "Browse media sources", Transport: "ws", Endpoint: "media_source/browse_media", Effect: "read", Args: []string{"media_content_id"}, OptionalArg: true},
		{Use: "search <search_query>", Short: "Search a media source", Transport: "ws", Endpoint: "media_source/search_media", Effect: "read", Args: []string{"search_query"}, Input: true},
		{Use: "resolve <media_content_id>", Short: "Resolve a playable media URL", Transport: "ws", Endpoint: "media_source/resolve_media", Effect: "read", Args: []string{"media_content_id"}},
		{Use: "delete <media_content_id>", Short: "Remove local media", Transport: "ws", Endpoint: "media_source/local_source/remove", Effect: "destructive", Args: []string{"media_content_id"}},
	})
}
