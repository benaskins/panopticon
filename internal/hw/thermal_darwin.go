//go:build darwin

package hw

/*
#cgo LDFLAGS: -framework Foundation
#include <stdlib.h>

extern int getThermalStateValue();
*/
import "C"

// ThermalState returns the current thermal state as a string.
func ThermalState() string {
	state := int(C.getThermalStateValue())
	switch state {
	case 0:
		return "Nominal"
	case 1:
		return "Fair"
	case 2:
		return "Serious"
	case 3:
		return "Critical"
	default:
		return "Unknown"
	}
}
