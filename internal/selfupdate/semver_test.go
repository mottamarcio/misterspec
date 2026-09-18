package selfupdate

import "testing"

func TestNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v1.1.5-alpha", "v1.2.0", true},
		{"v1.2.0", "v1.2.0", false},
		{"v1.2.0", "v1.1.5-alpha", false},
		{"development build", "v1.2.0", false},
	}
	for _, c := range cases {
		if got := Newer(c.current, c.latest); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}
