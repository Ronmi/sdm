package sdm

import (
	"errors"
	"strings"

	"git.ronmi.tw/ronmi/sdm/driver"
)

var registeredDrivers = map[string]driver.DriverFactory{}

// RegisterDriver registers a driver factory for use in sdm.
//
// Note: later one with the same name will be discarded.
func RegisterDriver(name string, f driver.DriverFactory) {
	if _, ok := registeredDrivers[name]; ok {
		return
	}

	registeredDrivers[name] = f
}

func getDriver(drvStr string) driver.Driver {
	idx := strings.Index(drvStr, ":")
	name := drvStr
	if idx != -1 {
		name = drvStr[:idx]
	}

	factory, ok := registeredDrivers[name]
	if !ok {
		panic(errors.New("sdm: driver " + name + " not found"))
	}

	params := map[string]string{}

	if idx != -1 {
		// parse params
		strs := strings.Split(name[idx+1:], ";")
		for _, paramStr := range strs {
			idx := strings.Index(paramStr, "=")
			params[paramStr[:idx]] = paramStr[idx+1:]
		}
	}

	return factory(params)
}
