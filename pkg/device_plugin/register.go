package device_plugin

import (
	"context"
	"github.com/lixd/i-device-plugin/pkg/common"
	"github.com/pkg/errors"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
)

// Register registers the device plugin for the given resourceName with Kubelet.
func (c *GopherDevicePlugin) Register() error {
	conn, err := connect(pluginapi.KubeletSocket, common.ConnectTimeout)
	if err != nil {
		return errors.WithMessagef(err, "connect to %s failed", pluginapi.KubeletSocket)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(common.DeviceSocket),
		ResourceName: common.ResourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return errors.WithMessage(err, "register to kubelet failed")
	}
	return nil
}
