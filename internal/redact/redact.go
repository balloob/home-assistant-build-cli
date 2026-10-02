// Package redact provides conservative, non-mutating redaction for plans.
package redact

import (
	"net/url"
	"strings"
)

const Hidden = "[REDACTED]"

func Map(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = Value(key, item)
	}
	return result
}

func Sensitive(key string) bool {
	key = strings.ReplaceAll(strings.ToLower(key), "-", "_")
	for _, word := range []string{"password", "passwd", "passphrase", "token", "secret", "credential", "authorization", "cookie", "private_key", "encryption_key", "api_key", "apikey"} {
		if strings.Contains(key, word) {
			return true
		}
	}
	parts := strings.Split(key, "/")
	field := parts[len(parts)-1]
	if field == "key" || field == "pin" || field == "psk" || strings.HasSuffix(field, "_pin") {
		return true
	}
	return false
}

func Value(key string, value any) any {
	if Sensitive(key) {
		return Hidden
	}
	switch v := value.(type) {
	case map[string]any:
		copy := make(map[string]any, len(v))
		for k, item := range v {
			copy[k] = Value(k, item)
		}
		return copy
	case []any:
		copy := make([]any, len(v))
		for i, item := range v {
			copy[i] = Value(key, item)
		}
		return copy
	case string:
		u, err := url.Parse(v)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return v
		}
		if u.User != nil {
			u.User = url.User(Hidden)
		}
		q := u.Query()
		for name := range q {
			if Sensitive(name) {
				q.Set(name, Hidden)
			}
		}
		u.RawQuery = q.Encode()
		return u.String()
	default:
		return value
	}
}
