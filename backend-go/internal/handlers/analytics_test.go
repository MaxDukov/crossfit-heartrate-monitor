package handlers

import (
	"testing"
	"time"
)

func TestParseDBTime(t *testing.T) {
	cases := []struct {
		in   string
		want int64 // unix
	}{
		{"2026-08-30 19:47:59.992458", 1788119279},
		{"2026-08-30 19:47:59", 1788119279},
		{"2026-08-30T19:47:59.992458Z", 1788119279},
	}
	for _, c := range cases {
		got := parseDBTime(c.in).Unix()
		if got != c.want {
			t.Errorf("parseDBTime(%q) = %d, want %d", c.in, got, c.want)
		}
	}
	if !parseDBTime("мусор").IsZero() {
		t.Error("мусор должен давать zero time")
	}
}

func TestRound1(t *testing.T) {
	if round1(84.46) != 84.5 {
		t.Errorf("round1(84.46) = %v", round1(84.46))
	}
	if round1(82.15) != 82.2 {
		t.Errorf("round1(82.15) = %v", round1(82.15))
	}
	_ = time.Now()
}
