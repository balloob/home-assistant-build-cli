package auth

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// DiscoveredServer represents a Home Assistant server found via mDNS
type DiscoveredServer struct {
	Name     string   // Instance name (e.g., "Home" or location name)
	Host     string   // Hostname
	Port     int      // Port number
	IPv4     []net.IP // IPv4 addresses
	IPv6     []net.IP // IPv6 addresses
	Version  string   // Home Assistant version (from TXT records)
	UUID     string   // Instance UUID (from TXT records)
	URL      string   // Constructed URL for connecting
	Internal string   // Internal URL (from TXT records)
	External string   // External URL (from TXT records)
}

// DiscoverServers searches for Home Assistant instances on the local network
// using mDNS/DNS-SD. It searches for the _home-assistant._tcp service type.
// The timeout parameter controls how long to search (recommended: 3-5 seconds).
func DiscoverServers(timeout time.Duration) ([]DiscoveredServer, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	servers := make([]DiscoveredServer, 0)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start browsing for Home Assistant services
	go func() {
		// Home Assistant broadcasts as _home-assistant._tcp
		err = resolver.Browse(ctx, "_home-assistant._tcp", "local.", entries)
		if err != nil {
			// Log error but don't fail - we might still get results
		}
	}()

	// Collect discovered entries, deduplicating by URL
	seen := make(map[string]int) // URL -> index in servers slice
	for entry := range entries {
		server := parseServiceEntry(entry)
		if server.URL == "" {
			continue
		}

		if existingIdx, exists := seen[server.URL]; exists {
			// Keep the entry with the shorter name (likely the original, not "Home-2", "Home-3", etc.)
			if len(server.Name) < len(servers[existingIdx].Name) {
				servers[existingIdx] = server
			}
		} else {
			seen[server.URL] = len(servers)
			servers = append(servers, server)
		}
	}

	// Sort by name for consistent display
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})

	return servers, nil
}

// parseServiceEntry converts a zeroconf entry to a DiscoveredServer
func parseServiceEntry(entry *zeroconf.ServiceEntry) DiscoveredServer {
	// Unescape the instance name (mDNS escapes spaces as "\ ")
	name := strings.ReplaceAll(entry.Instance, `\ `, " ")

	server := DiscoveredServer{
		Name: name,
		Host: entry.HostName,
		Port: entry.Port,
		IPv4: entry.AddrIPv4,
		IPv6: entry.AddrIPv6,
	}

	// Parse TXT records for additional metadata
	for _, txt := range entry.Text {
		if len(txt) > 8 && txt[:8] == "version=" {
			server.Version = txt[8:]
		} else if len(txt) > 5 && txt[:5] == "uuid=" {
			server.UUID = txt[5:]
		} else if len(txt) > 13 && txt[:13] == "internal_url=" {
			server.Internal = txt[13:]
		} else if len(txt) > 13 && txt[:13] == "external_url=" {
			server.External = txt[13:]
		}
	}

	// Construct URL - prefer internal URL from TXT, then IPv4, then IPv6
	server.URL = constructURL(server)

	return server
}

// constructURL builds a connection URL for the discovered server
func constructURL(server DiscoveredServer) string {
	// Prefer internal URL if provided in TXT records
	if server.Internal != "" {
		return server.Internal
	}

	// Determine protocol based on port
	protocol := "http"
	if server.Port == 443 || server.Port == 8443 {
		protocol = "https"
	}

	// Prefer IPv4 addresses
	if len(server.IPv4) > 0 {
		addr := server.IPv4[0].String()
		if server.Port == 80 || server.Port == 443 {
			return fmt.Sprintf("%s://%s", protocol, addr)
		}
		return fmt.Sprintf("%s://%s:%d", protocol, addr, server.Port)
	}

	// Fall back to IPv6
	if len(server.IPv6) > 0 {
		addr := server.IPv6[0].String()
		if server.Port == 80 || server.Port == 443 {
			return fmt.Sprintf("%s://[%s]", protocol, addr)
		}
		return fmt.Sprintf("%s://[%s]:%d", protocol, addr, server.Port)
	}

	// Last resort: use hostname
	if server.Host != "" {
		host := server.Host
		// Remove trailing dot from DNS hostname
		if host[len(host)-1] == '.' {
			host = host[:len(host)-1]
		}
		if server.Port == 80 || server.Port == 443 {
			return fmt.Sprintf("%s://%s", protocol, host)
		}
		return fmt.Sprintf("%s://%s:%d", protocol, host, server.Port)
	}

	return ""
}

// FormatServerDisplay returns a human-readable string for displaying a server option
func FormatServerDisplay(server DiscoveredServer) string {
	display := server.Name
	if server.Version != "" {
		display = fmt.Sprintf("%s (v%s)", display, server.Version)
	}
	if server.URL != "" {
		display = fmt.Sprintf("%s - %s", display, server.URL)
	}
	return display
}
