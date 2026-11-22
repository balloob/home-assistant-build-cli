package client

import (
	"context"
	"strings"
)

func (c *RestClient) RequestContext(ctx context.Context, method, endpoint string, body any) (any, error) {
	req := c.getClient().R().SetContext(ctx)
	if strings.HasPrefix(endpoint, "config/config_entries/") {
		// OAuth flows return the user to the Home Assistant frontend.
		req.SetHeader("HA-Frontend-Base", strings.TrimRight(c.BaseURL, "/"))
	}
	if body != nil {
		req.SetBody(body)
	}
	response, err := req.Execute(method, "/api/"+endpoint)
	if err != nil {
		return nil, err
	}
	return c.handleResponse(response)
}
