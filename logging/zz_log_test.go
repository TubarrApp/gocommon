package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
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

// TestChannelLogger tests that a channel logger's lines carry the channel, others don't, and caller info still names
// the function that logged.
func TestChannelLogger(t *testing.T) {
	pl, err := SetupLogging(Config{LogFilePath: filepath.Join(t.TempDir(), "test.log"), Console: io.Discard, Program: "ChannelTest"})
	if err != nil {
		t.Fatal(err)
	}
	_, seen := pl.LogsSince(0)

	pl.Channel("Cats").E("crawl failed")
	pl.I("not about a channel")
	lines, _ := pl.LogsSince(seen)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}

	var tagged, plain map[string]any
	if err := json.Unmarshal(lines[0], &tagged); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(lines[1], &plain); err != nil {
		t.Fatal(err)
	}
	if tagged[jChannel] != "Cats" {
		t.Errorf("channel = %v, want Cats", tagged[jChannel])
	}
	if fn, _ := tagged[jFunction].(string); !strings.HasSuffix(fn, "TestChannelLogger") {
		t.Errorf("function = %q, want the test that logged", fn)
	}
	if _, ok := plain[jChannel]; ok {
		t.Errorf("untagged line has a channel: %s", lines[1])
	}
}
