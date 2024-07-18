package common

import "time"

const (
	ResourceName   string = "lixueduan.com/gopher"
	DevicePath     string = "/etc/gophers"
	DeviceSocket   string = "gopher.sock"
	ConnectTimeout        = time.Second * 5
)
