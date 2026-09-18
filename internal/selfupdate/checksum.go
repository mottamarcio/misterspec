package selfupdate

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ParseChecksums parses a SHA256SUMS-format file (one
// "<64-hex-digest>  <filename>" line per asset — contracts/version-update.md)
// into a filename → lowercase hex digest map.
func ParseChecksums(data []byte) (map[string]string, error) {
	out := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("selfupdate: malformed SHA256SUMS line: %q", line)
		}
		out[fields[1]] = strings.ToLower(fields[0])
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// VerifyChecksum reports whether content's own SHA-256 matches
// expectedHex (case-insensitive) — the gate a downloaded binary MUST
// pass before it is ever used to replace anything (FR-007/FR-008).
func VerifyChecksum(content []byte, expectedHex string) bool {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]) == strings.ToLower(expectedHex)
}
