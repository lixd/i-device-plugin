package utils

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// TestWatchKubelet test scenario: simulate kubelet restart (delete and recreate kubelet.sock)
func TestWatchKubelet(t *testing.T) {
	// Create temporary directory to simulate DevicePluginPath
	tmpDir, err := ioutil.TempDir("", "device-plugin")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to modify DevicePluginPath temporarily
	kubeletSock := tmpDir + "/kubelet.sock"
	os.Setenv("KUBELET_SOCKET", kubeletSock)
	defer os.Unsetenv("KUBELET_SOCKET")

	// Create initial kubelet.sock file
	if err := ioutil.WriteFile(kubeletSock, []byte{}, 0666); err != nil {
		t.Fatalf("failed to create initial kubelet.sock file: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create fsnotify watcher: %v", err)
	}
	defer watcher.Close()

	stop := make(chan struct{})

	if err := WatchKubelet(watcher, stop); err != nil {
		t.Fatalf("WatchKubelet failed: %v", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to ensure watcher is ready

		// Remove kubelet.sock
		if err := os.Remove(kubeletSock); err != nil {
			t.Errorf("failed to remove kubelet.sock file: %v", err)
			return
		}

		// Short delay before recreating kubelet.sock
		time.Sleep(100 * time.Millisecond)
		if err := ioutil.WriteFile(kubeletSock, []byte{}, 0666); err != nil {
			t.Errorf("failed to recreate kubelet.sock file: %v", err)
		}
	}()

	select {
	case <-stop:
		// Successfully received restart signal
		t.Log("successfully received restart signal")
	case <-time.After(5 * time.Second):
		t.Error("timeout: did not receive expected restart signal")
	}
}
