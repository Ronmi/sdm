package sdm

import "git.ronmi.tw/ronmi/sdm/driver"

var registeredDrivers = map[string]driver.Driver{}

// RegisterDriver registers a driver for use in sdm.
//
// Note: later one with the same name will be discarded.
func RegisterDriver(name string, drv driver.Driver) {
	if _, ok := registeredDrivers[name]; ok {
		return
	}

	registeredDrivers[name] = drv
}
