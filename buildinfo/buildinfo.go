// Package buildinfo exposes information about the build for the running binary.
package buildinfo

import (
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// Build setting keys stamped in by the Go toolchain.
const (
	vcsRevision = "vcs.revision"
	vcsTime     = "vcs.time"
	vcsModified = "vcs.modified"
)

// Formatting constants for Info.String.
const (
	timeFormat = "2006-01-02 15:04:05 MST"
	unknown    = "unknown"
	shortHash  = 7
)

// Info describes how and when the running binary was built.
type Info struct {
	Revision  string    // Full commit hash, empty if unavailable.
	Committed time.Time // Commit timestamp, zero if unavailable.
	Built     time.Time // Binary modification time, zero if unavailable.
	Modified  bool      // True if built from a dirty working tree.
	GoVersion string    // Toolchain used to compile.
}

// Get reads build provenance from the embedded build info and the executable itself.
func Get() Info {
	var i Info

	if bi, ok := debug.ReadBuildInfo(); ok {
		i.GoVersion = bi.GoVersion

		for _, s := range bi.Settings {
			switch s.Key {
			case vcsRevision:
				i.Revision = s.Value
			case vcsTime:
				if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
					i.Committed = t.Local()
				}
			case vcsModified:
				i.Modified, _ = strconv.ParseBool(s.Value)
			}
		}
	}

	if exe, err := os.Executable(); err == nil {
		if fi, err := os.Stat(exe); err == nil {
			i.Built = fi.ModTime()
		}
	}

	return i
}

// ShortRevision returns the abbreviated commit hash, or an empty string if unavailable.
func (i Info) ShortRevision() string {
	if len(i.Revision) > shortHash {
		return i.Revision[:shortHash]
	}
	return i.Revision
}

// String renders the build info as a single human readable line.
//
// For example: "built 2026-09-05 22:47:10 BST (commit 270d833 of 2026-09-05 22:31:40 BST)".
func (i Info) String() string {
	var b strings.Builder

	b.WriteString("built ")
	if i.Built.IsZero() {
		b.WriteString(unknown)
	} else {
		b.WriteString(i.Built.Format(timeFormat))
	}

	rev := i.ShortRevision()
	if rev == "" {
		return b.String()
	}

	b.WriteString(" (commit ")
	b.WriteString(rev)

	if !i.Committed.IsZero() {
		b.WriteString(" of ")
		b.WriteString(i.Committed.Format(timeFormat))
	}
	if i.Modified {
		b.WriteString(", dirty")
	}
	b.WriteString(")")

	return b.String()
}
