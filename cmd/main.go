package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/lixd/i-device-plugin/pkg/device_plugin"
	"github.com/lixd/i-device-plugin/pkg/utils"
	"k8s.io/klog/v2"
)

func main() {
	klog.Infof("device plugin starting")
	dp := device_plugin.NewGopherDevicePlugin()
	go dp.Run()

	// register when device plugin start
	if err := dp.Register(); err != nil {
		klog.Fatalf("register to kubelet failed: %v", err)
	}

	// watch kubelet.sock,when kubelet restart,exit device plugin,then will restart by DaemonSet
	stop := make(chan struct{})
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Fatalf("Unable to create fsnotify watcher: %v", err)
	}
	defer watcher.Close()

	if err = utils.WatchKubelet(watcher, stop); err != nil {
		klog.Fatalf("start to kubelet failed: %v", err)
	}

	<-stop
	klog.Infof("kubelet restart,exiting")
}
