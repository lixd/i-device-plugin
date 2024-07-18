package utils

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// WatchKubelet restart device plugin when kubelet restarted
func WatchKubelet(stop chan<- struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithMessage(err, "Unable to create fsnotify watcher")
	}
	defer watcher.Close()

	go func() {
		// Start listening for events.
		for {
			select {
			case event := <-watcher.Events:
				klog.Infof("fsnotify events: %s %v", event.Name, event.Op.String())
				if event.Name == pluginapi.KubeletSocket && event.Op == fsnotify.Create {
					klog.Warning("inotify: kubelet.sock created, restarting.")
					stop <- struct{}{}
				}
			case err = <-watcher.Errors:
				klog.Errorf("fsnotify failed restarting,detail:%v", err)
			}
		}
	}()

	// watch kubelet.sock
	err = watcher.Add(pluginapi.KubeletSocket)
	if err != nil {
		return errors.WithMessagef(err, "Unable to add path %s to watcher", pluginapi.KubeletSocket)
	}
	return nil
}
