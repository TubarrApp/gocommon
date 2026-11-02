//go:build !windows

package sharedfiles

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestOpenRegularInRootRefusesFIFO tests a FIFO is refused, rather than blocking until something writes to it.
func TestOpenRegularInRootRefusesFIFO(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "pipe.json")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		f, err := OpenRegularInRoot(fifo)
		if f != nil {
			_ = f.Close()
		}
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected an error for a FIFO")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("OpenRegularInRoot blocked on a FIFO")
	}
}

// TestOpenRegularInRoot tests a regular file still opens.
func TestOpenRegularInRoot(t *testing.T) {
	p := filepath.Join(t.TempDir(), "meta.json")
	if err := os.WriteFile(p, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	f, err := OpenRegularInRoot(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Error(err)
	}
}
