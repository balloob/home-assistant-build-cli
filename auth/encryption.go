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
	"slices"
	"strings"
	"sync"
)

// getMachineID returns a machine-specific identifier for key derivation.
// It is the platform machine ID alone, or the hostname where there is none.
func getMachineID() string {
	if hwUUID := getHardwareUUID(); hwUUID != "" {
		return hwUUID
	}
	hostname, _ := os.Hostname()
	return hostname
}

// legacyMachineIDs returns the identifiers older versions derived the key
// from: hostname, the MAC of the first active network interface, and the
// machine ID. That interface changes with VPNs, Hyper-V/WSL/Docker adapters
// and Wi-Fi, so there is one candidate per MAC on the machine, and one
// without a MAC.
func legacyMachineIDs() []string {
	macs := []string{""}
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			if mac := iface.HardwareAddr.String(); mac != "" && !slices.Contains(macs, mac) {
				macs = append(macs, mac)
			}
		}
	}

	hostname, hostnameErr := os.Hostname()
	hwUUID := getHardwareUUID()
	ids := make([]string, 0, len(macs))
	for _, mac := range macs {
		var sources []string
		if hostnameErr == nil {
			sources = append(sources, hostname)
		}
		if mac != "" {
			sources = append(sources, mac)
		}
		if hwUUID != "" {
			sources = append(sources, hwUUID)
		}
		ids = append(ids, strings.Join(sources, "|"))
	}
	return ids
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
		hash := sha256.Sum256([]byte(getMachineID()))
		derivedKey = hash[:]
	})
	return derivedKey
}

// legacyKeys returns the old hostname- and MAC-dependent keys, used only to
// read files written by older versions.
func legacyKeys() [][]byte {
	ids := legacyMachineIDs()
	keys := make([][]byte, 0, len(ids))
	for _, id := range ids {
		hash := sha256.Sum256([]byte(id))
		keys = append(keys, hash[:])
	}
	return keys
}

// newGCM creates an AES-GCM cipher from the given key.
func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	return gcm, nil
}

// encrypt encrypts data using AES-GCM
func encrypt(plaintext, key []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
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
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
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
