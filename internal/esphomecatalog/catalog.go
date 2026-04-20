package esphomecatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type githubYAMLRef struct {
	Owner string
	Repo  string
	Ref   string
	Path  string
}

const (
	defaultAPIRoot = "https://api.github.com"
	defaultRawRoot = "https://raw.githubusercontent.com"
	defaultOwner   = "esphome"
	defaultRepo    = "esphome-devices"
	defaultRef     = "main"
)

// SearchOptions controls catalog search behavior.
type SearchOptions struct {
	Query      string
	Board      string
	DeviceType string
	Difficulty int
	Limit      int
}

// Entry is a lightweight catalog result.
type Entry struct {
	Slug       string `json:"slug"`
	Title      string `json:"title,omitempty"`
	Board      string `json:"board,omitempty"`
	DeviceType string `json:"type,omitempty"`
	Difficulty int    `json:"difficulty,omitempty"`
	ProjectURL string `json:"project_url,omitempty"`
}

// Device is a detailed catalog record.
type Device struct {
	Entry
	ProjectName      string `json:"project_name,omitempty"`
	PackageImportURL string `json:"package_import_url,omitempty"`
	YAMLPath         string `json:"yaml_path,omitempty"`
	YAMLContent      string `json:"yaml_content,omitempty"`
	ContentSource    string `json:"content_source,omitempty"`
	MarkdownPath     string `json:"markdown_path,omitempty"`
}

type repoContent struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

// Client fetches and parses devices from the ESPHome catalog repo.
type Client struct {
	httpClient *http.Client
	apiRoot    string
	rawRoot    string
	owner      string
	repo       string
	ref        string
	token      string
}

var (
	lookPath       = exec.LookPath
	ghTokenCommand = func(ctx context.Context, path string) ([]byte, error) {
		return exec.CommandContext(ctx, path, "auth", "token").Output()
	}
)

// NewClient creates a catalog client for a git ref.
func NewClient(ref string) *Client {
	if strings.TrimSpace(ref) == "" {
		ref = defaultRef
	}

	token := discoverGitHubToken()

	return &Client{
		httpClient: &http.Client{Timeout: 20 * time.Second},
		apiRoot:    defaultAPIRoot,
		rawRoot:    defaultRawRoot,
		owner:      defaultOwner,
		repo:       defaultRepo,
		ref:        ref,
		token:      token,
	}
}

func discoverGitHubToken() string {
	token := strings.TrimSpace(os.Getenv("GH_TOKEN"))
	if token != "" {
		return token
	}
	token = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token != "" {
		return token
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	path, err := lookPath("gh")
	if err != nil || path == "" {
		return ""
	}
	out, err := ghTokenCommand(ctx, path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Search finds devices by slug/title and optional metadata filters.
func (c *Client) Search(ctx context.Context, options SearchOptions) ([]Entry, error) {
	query := strings.ToLower(strings.TrimSpace(options.Query))
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	limit := options.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	slugs, err := c.listDeviceSlugs(ctx)
	if err != nil {
		return nil, err
	}

	boardFilter := strings.ToLower(strings.TrimSpace(options.Board))
	typeFilter := strings.ToLower(strings.TrimSpace(options.DeviceType))

	results := make([]Entry, 0, limit)
	for _, slug := range slugs {
		if !strings.Contains(strings.ToLower(slug), query) {
			continue
		}

		device, err := c.GetDevice(ctx, slug)
		if err != nil {
			continue
		}

		if boardFilter != "" && !containsToken(device.Board, boardFilter) {
			continue
		}
		if typeFilter != "" && strings.ToLower(device.DeviceType) != typeFilter {
			continue
		}
		if options.Difficulty > 0 && device.Difficulty != options.Difficulty {
			continue
		}

		results = append(results, device.Entry)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// GetDevice fetches and parses a specific device page.
func (c *Client) GetDevice(ctx context.Context, slug string) (*Device, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	markdownPath, markdown, err := c.readDeviceMarkdown(ctx, slug)
	if err != nil {
		return nil, err
	}

	frontmatter := parseFrontmatter(markdown)
	device := &Device{
		Entry: Entry{
			Slug:       slug,
			Title:      stringField(frontmatter, "title"),
			Board:      stringField(frontmatter, "board"),
			DeviceType: stringField(frontmatter, "type"),
			Difficulty: intField(frontmatter, "difficulty"),
			ProjectURL: stringField(frontmatter, "project-url"),
		},
		ProjectName:      stringField(frontmatter, "project-name"),
		PackageImportURL: normalizePackageImportURL(stringField(frontmatter, "package_import_url")),
		MarkdownPath:     markdownPath,
	}

	if device.Title == "" {
		device.Title = slug
	}
	if device.ProjectName == "" {
		device.ProjectName = "esphome." + normalizeSlug(slug)
	}

	if device.PackageImportURL == "" {
		if projectURL := strings.TrimSpace(device.ProjectURL); isYAMLURL(projectURL) {
			device.PackageImportURL = normalizePackageImportURL(projectURL)
		}
	}

	if path, content, err := c.readLinkedYAML(ctx, slug, markdown); err == nil && strings.TrimSpace(content) != "" {
		device.YAMLPath = path
		device.YAMLContent = content
		device.ContentSource = "yaml_file"
		return device, nil
	}

	if yamlBlock := firstYAMLBlock(markdown); yamlBlock != "" {
		device.YAMLPath = markdownPath + "#codeblock"
		device.YAMLContent = yamlBlock
		device.ContentSource = "yaml_block"
	}

	return device, nil
}

func (c *Client) listDeviceSlugs(ctx context.Context) ([]string, error) {
	path := fmt.Sprintf("/repos/%s/%s/contents/src/docs/devices", c.owner, c.repo)
	var items []repoContent
	if err := c.apiGetJSON(ctx, path, map[string]string{"ref": c.ref}, &items); err != nil {
		return nil, err
	}

	slugs := make([]string, 0, len(items))
	for _, item := range items {
		if item.Type == "dir" {
			slugs = append(slugs, item.Name)
		}
	}

	sort.Strings(slugs)
	return slugs, nil
}

func (c *Client) readDeviceMarkdown(ctx context.Context, slug string) (string, string, error) {
	indexPath := fmt.Sprintf("src/docs/devices/%s/index.md", slug)
	text, err := c.rawGet(ctx, indexPath)
	if err == nil {
		return indexPath, text, nil
	}

	path := fmt.Sprintf("/repos/%s/%s/contents/src/docs/devices/%s", c.owner, c.repo, url.PathEscape(slug))
	var items []repoContent
	if apiErr := c.apiGetJSON(ctx, path, map[string]string{"ref": c.ref}, &items); apiErr != nil {
		return "", "", err
	}

	markdownFile := ""
	for _, item := range items {
		if item.Type == "file" && strings.HasSuffix(item.Name, ".md") {
			if markdownFile != "" {
				return "", "", fmt.Errorf("multiple markdown files found for catalog device %q", slug)
			}
			markdownFile = item.Name
		}
	}

	if markdownFile == "" {
		return "", "", fmt.Errorf("no markdown file found for catalog device %q", slug)
	}

	markdownPath := fmt.Sprintf("src/docs/devices/%s/%s", slug, markdownFile)
	markdown, err := c.rawGet(ctx, markdownPath)
	if err != nil {
		return "", "", err
	}

	return markdownPath, markdown, nil
}

func (c *Client) readLinkedYAML(ctx context.Context, slug, markdown string) (string, string, error) {
	link := firstYAMLLink(markdown)
	if link == "" {
		return "", "", fmt.Errorf("no yaml link found")
	}

	if isHTTPURL(link) {
		link = normalizeLinkedYAMLURL(c.rawRoot, link)
		content, err := c.getURL(ctx, link)
		if err != nil {
			return "", "", err
		}
		return link, content, nil
	}

	path := fmt.Sprintf("src/docs/devices/%s/%s", slug, strings.TrimLeft(link, "/"))
	content, err := c.rawGet(ctx, path)
	if err != nil {
		return "", "", err
	}
	return path, content, nil
}

func (c *Client) apiGetJSON(ctx context.Context, path string, query map[string]string, target any) error {
	base := strings.TrimRight(c.apiRoot, "/")
	u, err := url.Parse(base + path)
	if err != nil {
		return err
	}

	values := u.Query()
	for key, value := range query {
		if value != "" {
			values.Set(key, value)
		}
	}
	u.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return formatGitHubHTTPError("catalog API", resp.StatusCode, body, c.token != "")
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("parse catalog API response: %w", err)
	}

	return nil
}

func (c *Client) rawGet(ctx context.Context, path string) (string, error) {
	path = strings.TrimLeft(path, "/")
	u := fmt.Sprintf("%s/%s/%s/%s/%s", strings.TrimRight(c.rawRoot, "/"), c.owner, c.repo, c.ref, path)
	return c.getURL(ctx, u)
}

func (c *Client) getURL(ctx context.Context, value string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, value, nil)
	if err != nil {
		return "", err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", formatGitHubHTTPError("catalog content", resp.StatusCode, body, c.token != "")
	}

	return string(body), nil
}

func formatGitHubHTTPError(prefix string, statusCode int, body []byte, authenticated bool) error {
	trimmed := strings.TrimSpace(string(body))
	if statusCode == http.StatusForbidden && strings.Contains(strings.ToLower(trimmed), "rate limit") {
		if authenticated {
			return fmt.Errorf("%s rate limit exceeded even with authentication: %s", prefix, trimmed)
		}
		return fmt.Errorf("%s rate limit exceeded; set GH_TOKEN/GITHUB_TOKEN or authenticate gh CLI to raise the limit: %s", prefix, trimmed)
	}
	return fmt.Errorf("%s returned status %d: %s", prefix, statusCode, trimmed)
}

func parseFrontmatter(markdown string) map[string]any {
	trimmed := strings.TrimLeft(markdown, "\ufeff")
	if !strings.HasPrefix(trimmed, "---\n") {
		return map[string]any{}
	}

	parts := strings.SplitN(trimmed, "\n---\n", 2)
	if len(parts) != 2 {
		return map[string]any{}
	}

	meta := map[string]any{}
	if err := yaml.Unmarshal([]byte(parts[0][4:]), &meta); err != nil {
		return map[string]any{}
	}

	return meta
}

func firstYAMLLink(markdown string) string {
	re := regexp.MustCompile(`\(([^)\s]+\.ya?ml)\)`)
	matches := re.FindAllStringSubmatch(markdown, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		link := strings.TrimSpace(match[1])
		if strings.Contains(strings.ToLower(link), "resources/") {
			continue
		}
		return link
	}
	return ""
}

func firstYAMLBlock(markdown string) string {
	re := regexp.MustCompile("(?s)```yaml\\s*(.*?)\\s*```")
	match := re.FindStringSubmatch(markdown)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1]) + "\n"
}

func stringField(meta map[string]any, key string) string {
	if value, ok := meta[key]; ok {
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}

	alias := strings.ReplaceAll(key, "-", "_")
	if value, ok := meta[alias]; ok {
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}

	return ""
}

func intField(meta map[string]any, key string) int {
	value, ok := meta[key]
	if !ok {
		alias := strings.ReplaceAll(key, "-", "_")
		value = meta[alias]
	}

	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}

	return 0
}

func normalizeSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return strings.ReplaceAll(slug, "-", "_")
}

func containsToken(raw, needle string) bool {
	raw = strings.ToLower(raw)
	needle = strings.ToLower(needle)
	for _, token := range strings.Split(raw, ",") {
		if strings.TrimSpace(token) == needle {
			return true
		}
	}
	return false
}

func isYAMLURL(value string) bool {
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}

func isHTTPURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func normalizePackageImportURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(value), "github://") {
		return value
	}

	ref, ok := parseGitHubYAMLRef(value)
	if !ok {
		return value
	}

	return fmt.Sprintf("github://%s/%s/%s@%s", ref.Owner, ref.Repo, ref.Path, ref.Ref)
}

func normalizeLinkedYAMLURL(rawRoot, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	ref, ok := parseGitHubYAMLRef(value)
	if !ok {
		return value
	}

	base := strings.TrimRight(rawRoot, "/")
	return fmt.Sprintf("%s/%s/%s/%s/%s", base, ref.Owner, ref.Repo, ref.Ref, ref.Path)
}

func parseGitHubYAMLRef(value string) (githubYAMLRef, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return githubYAMLRef{}, false
	}

	if strings.HasPrefix(strings.ToLower(value), "github://") {
		trimmed := strings.TrimPrefix(value, "github://")
		owner, rest, ok := strings.Cut(trimmed, "/")
		if !ok {
			return githubYAMLRef{}, false
		}
		repo, pathWithRef, ok := strings.Cut(rest, "/")
		if !ok {
			return githubYAMLRef{}, false
		}
		at := strings.LastIndex(pathWithRef, "@")
		if at <= 0 || at == len(pathWithRef)-1 {
			return githubYAMLRef{}, false
		}
		ref := strings.TrimSpace(pathWithRef[at+1:])
		path := strings.Trim(pathWithRef[:at], "/")
		if owner == "" || repo == "" || ref == "" || !isYAMLURL(path) {
			return githubYAMLRef{}, false
		}
		return githubYAMLRef{Owner: owner, Repo: repo, Ref: ref, Path: path}, true
	}

	u, err := url.Parse(value)
	if err != nil {
		return githubYAMLRef{}, false
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	switch strings.ToLower(u.Host) {
	case "github.com":
		if len(segments) < 5 || segments[2] != "blob" {
			return githubYAMLRef{}, false
		}
		path := strings.Join(segments[4:], "/")
		if !isYAMLURL(path) {
			return githubYAMLRef{}, false
		}
		return githubYAMLRef{
			Owner: segments[0],
			Repo:  segments[1],
			Ref:   segments[3],
			Path:  path,
		}, true
	case "raw.githubusercontent.com":
		if len(segments) < 4 {
			return githubYAMLRef{}, false
		}
		path := strings.Join(segments[3:], "/")
		if !isYAMLURL(path) {
			return githubYAMLRef{}, false
		}
		return githubYAMLRef{
			Owner: segments[0],
			Repo:  segments[1],
			Ref:   segments[2],
			Path:  path,
		}, true
	default:
		return githubYAMLRef{}, false
	}
}
