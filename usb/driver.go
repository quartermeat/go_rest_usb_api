package usb

import (
	"fmt"

	"github.com/google/gousb"
)

func FindDeviceWithVIDPID(vid, pid int) (*gousb.Device, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Iterate through all USB devices and find the one with the specified VID and PID.
	// This assumes that there is only one device with the specified VID and PID.
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(vid) && desc.Product == gousb.ID(pid)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open USB devices: %v", err)
	}
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	if len(devs) == 0 {
		return nil, fmt.Errorf("no USB devices found with VID 0x%x and PID 0x%x", vid, pid)
	}

	// Return the first matching device.
	return devs[0], nil
}
