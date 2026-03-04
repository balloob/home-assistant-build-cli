package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// getMachineID returns a machine-specific identifier for key derivation
func getMachineID() string {
	var sources []string

	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		sources = append(sources, hostname)
	}

	// MAC address
	if mac := getMACAddress(); mac != "" {
		sources = append(sources, mac)
	}

	// Platform-specific hardware UUID
	if hwUUID := getHardwareUUID(); hwUUID != "" {
		sources = append(sources, hwUUID)
	}

	return strings.Join(sources, "|")
}

func getMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			if mac := iface.HardwareAddr.String(); mac != "" {
				return mac
			}
		}
	}
	return ""
}

func getHardwareUUID() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: use hardware UUID
		cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		output, err := cmd.Output()
		if err != nil {
			return ""
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "IOPlatformUUID") {
				parts := strings.Split(line, "\"")
				if len(parts) >= 4 {
					return parts[3]
				}
			}
		}
	case "linux":
		// Linux: try /etc/machine-id
		if data, err := os.ReadFile("/etc/machine-id"); err == nil {
			return strings.TrimSpace(string(data))
		}
		// Fallback to /var/lib/dbus/machine-id
		if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
			return strings.TrimSpace(string(data))
		}
	case "windows":
		// Windows: use MachineGuid from registry
		cmd := exec.Command("reg", "query", "HKLM\\SOFTWARE\\Microsoft\\Cryptography", "/v", "MachineGuid")
		output, err := cmd.Output()
		if err != nil {
			return ""
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "MachineGuid") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return parts[len(parts)-1]
				}
			}
		}
	}
	return ""
}

// deriveKey derives an encryption key from machine-specific data.
// The result is cached after the first call because the underlying
// getMachineID() may spawn subprocesses (ioreg on macOS, reg query on
// Windows) which are expensive to repeat within a single CLI invocation.
var (
	derivedKey     []byte
	derivedKeyOnce sync.Once
)

func deriveKey() []byte {
	derivedKeyOnce.Do(func() {
		machineID := getMachineID()
		hash := sha256.Sum256([]byte(machineID))
		derivedKey = hash[:]
	})
	return derivedKey
}

// encrypt encrypts data using AES-GCM
func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
