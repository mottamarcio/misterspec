package selfupdate

import "testing"

func TestReport(t *testing.T) {
	t.Cleanup(func() { Version = "" })

	Version = "v1.2.0"
	if got := Report(); got != "v1.2.0" {
		t.Fatalf("Report() = %q, want %q", got, "v1.2.0")
	}

	Version = ""
	if got := Report(); got != "development build" {
		t.Fatalf("Report() = %q, want %q", got, "development build")
	}
}
