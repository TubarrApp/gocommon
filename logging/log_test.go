package logging

import (
	"fmt"
	"testing"
)

// TestLogsSince tests that only new lines are returned, across wraps, and that lines lost to the ring are skipped.
func TestLogsSince(t *testing.T) {
	pl := &ProgramLogger{logBuffer: make([][]byte, 3)}
	add := func(n int) {
		for range n {
			pl.addToRAMLine(fmt.Appendf(nil, "%d", pl.total))
		}
	}
	want := func(lines [][]byte, w ...string) {
		t.Helper()
		got := make([]string, len(lines))
		for i, l := range lines {
			got[i] = string(l)
		}
		if fmt.Sprint(got) != fmt.Sprint(w) {
			t.Fatalf("lines = %v, want %v", got, w)
		}
	}

	add(2)
	lines, next := pl.LogsSince(0)
	want(lines, "0", "1")

	add(2)
	lines, next = pl.LogsSince(next)
	want(lines, "2", "3")

	lines, next = pl.LogsSince(next)
	want(lines)

	add(5) // More than the ring holds: lines 4 and 5 are lost.
	lines, _ = pl.LogsSince(next)
	want(lines, "6", "7", "8")

	want(pl.GetRecentLogs(), "6", "7", "8")
}
