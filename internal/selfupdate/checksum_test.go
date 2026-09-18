package selfupdate

import "testing"

const sums = `d0be2dc421be4fcd0172e5afceea3970e2f3d940d0c6f37f0f9f18d9a8a24c1  misterspec-linux-amd64
b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944  misterspec-darwin-arm64
`

func TestParseChecksums(t *testing.T) {
	m, err := ParseChecksums([]byte(sums))
	if err != nil {
		t.Fatalf("ParseChecksums() error = %v", err)
	}
	if m["misterspec-linux-amd64"] != "d0be2dc421be4fcd0172e5afceea3970e2f3d940d0c6f37f0f9f18d9a8a24c1" {
		t.Errorf("m[misterspec-linux-amd64] = %q", m["misterspec-linux-amd64"])
	}
	if len(m) != 2 {
		t.Errorf("len(m) = %d, want 2", len(m))
	}
}

func TestVerifyChecksum(t *testing.T) {
	content := []byte("hello")
	const wantHex = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	if !VerifyChecksum(content, wantHex) {
		t.Errorf("VerifyChecksum(hello, %q) = false, want true", wantHex)
	}
	const zeroes = "0000000000000000000000000000000000000000000000000000000000000000"
	if VerifyChecksum(content, zeroes[:64]) {
		t.Error("VerifyChecksum(hello, zeros) = true, want false")
	}
}
