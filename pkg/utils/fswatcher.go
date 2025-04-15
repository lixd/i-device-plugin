package utils

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// WatchKubelet restart device plugin when kubelet restarted
func WatchKubelet(watcher *fsnotify.Watcher, stop chan<- struct{}) error {
	// Read path from environment variable first (for testing)
	kubeletSocket := os.Getenv("KUBELET_SOCKET")
	if kubeletSocket == "" {
		kubeletSocket = pluginapi.KubeletSocket
	}

	// Get directory path
	devicePluginPath := filepath.Dir(kubeletSocket)

	// watch dir /var/lib/kubelet/device-plugins/
	err := watcher.Add(devicePluginPath)
	if err != nil {
		return errors.WithMessagef(err, "Unable to add path %s to watcher", kubeletSocket)
	}

	go func() {
		// Start listening for events.
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				klog.Infof("fsnotify events: %s %v", event.Name, event.Op.String())
				if event.Name == kubeletSocket && event.Op == fsnotify.Create {
					klog.Warning("inotify: kubelet.sock created, restarting.")
					stop <- struct{}{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Errorf("fsnotify failed restarting,detail:%v", err)
			}
		}
	}()

	return nil
}
