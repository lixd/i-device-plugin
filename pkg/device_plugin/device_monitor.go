package device_plugin

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"io/fs"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
	"path/filepath"
	"strings"
)

type DeviceMonitor struct {
	path    string
	devices map[string]*pluginapi.Device
	notify  chan struct{} // notify when device update
}

func NewDeviceMonitor(path string) *DeviceMonitor {
	return &DeviceMonitor{
		path:    path,
		devices: make(map[string]*pluginapi.Device),
		notify:  make(chan struct{}),
	}
}

// List all device
func (d *DeviceMonitor) List() error {
	err := filepath.Walk(d.path, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			klog.Infof("%s is dir,skip", path)
			return nil
		}

		d.devices[info.Name()] = &pluginapi.Device{
			ID:     info.Name(),
			Health: pluginapi.Healthy,
		}
		return nil
	})

	return errors.WithMessagef(err, "walk [%s] failed", d.path)
}

// Watch device change
func (d *DeviceMonitor) Watch() error {
	klog.Infoln("watching devices")

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithMessage(err, "new watcher failed")
	}
	defer w.Close()

	errChan := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("device watcher panic:%v", r)
			}
		}()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				klog.Infof("fsnotify device event: %s %s", event.Name, event.Op.String())

				if event.Op == fsnotify.Create {
					dev := path.Base(event.Name)
					d.devices[dev] = &pluginapi.Device{
						ID:     dev,
						Health: pluginapi.Healthy,
					}
					d.notify <- struct{}{}
					klog.Infof("find new device [%s]", dev)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					dev := path.Base(event.Name)
					delete(d.devices, dev)
					d.notify <- struct{}{}
					klog.Infof("device [%s] removed", dev)
				}

			case err, ok := <-w.Errors:
				if !ok {
					continue
				}
				klog.Errorf("fsnotify watch device failed:%v", err)
			}
		}
	}()

	err = w.Add(d.path)
	if err != nil {
		return fmt.Errorf("watch device error:%v", err)
	}

	return <-errChan
}

// Devices transformer map to slice
func (d *DeviceMonitor) Devices() []*pluginapi.Device {
	devices := make([]*pluginapi.Device, 0, len(d.devices))
	for _, device := range d.devices {
		devices = append(devices, device)
	}
	return devices
}

func String(devs []*pluginapi.Device) string {
	ids := make([]string, 0, len(devs))
	for _, device := range devs {
		ids = append(ids, device.ID)
	}
	return strings.Join(ids, ",")
}
