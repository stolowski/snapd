// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/hotplug"
	"github.com/snapcore/snapd/interfaces/seccomp"
	"github.com/snapcore/snapd/snap"
)

const limeSdrSummary = `allows accessing Lime SDR`

const limeSdrBaseDeclarationSlots = `
  lime-sdr:
    allow-installation:
      slot-snap-type:
        - core
    deny-auto-connection: true
`

const limeSdrConnectedPlugApparmor = `
# for receiving kobject_uevent() net messages from the kernel
network netlink raw,

# Allow detection of usb devices. Leaks plugged in USB device info
/sys/bus/usb/devices/ r,

/run/udev/data/###MAJOR###:###MINOR### r,
/run/udev/data/+usb:* r,

# for read/write access to specific usb device
###USB_DEVICE### rw,

###SYSFS_PATH### r,
###SYSFS_PATH###/** r,

`

const limeSdrConnectedPlugSeccomp = `
# Description: Can access Unity7. Note, Unity 7 runs on X and requires access
# to various DBus services and this environment does not prevent eavesdropping
# or apps interfering with one another.

# Needed by QtSystems on X to detect mouse and keyboard
socket AF_NETLINK - NETLINK_KOBJECT_UEVENT
bind
`

var limeSdrProducts = []string{"04b4/8613/", "04b4/00f1/", "0403/601f/", "1d50/6108/"}

type limeSdrInterface struct{}

func (iface *limeSdrInterface) Name() string {
	return "lime-sdr"
}

func (iface *limeSdrInterface) StaticInfo() interfaces.StaticInfo {
	return interfaces.StaticInfo{
		Summary:              limeSdrSummary,
		BaseDeclarationSlots: limeSdrBaseDeclarationSlots,
	}
}

func (iface *limeSdrInterface) String() string {
	return iface.Name()
}

func (iface *limeSdrInterface) HotplugDeviceDetected(di *hotplug.HotplugDeviceInfo, spec *hotplug.Specification) error {
	if di.Subsystem() != "usb" {
		return nil
	}
	if devtype, ok := di.Attribute("DEVTYPE"); !ok || devtype != "usb_device" {
		return nil
	}
	if model, ok := di.Attribute("ID_MODEL"); ok && strings.HasPrefix(model, "LimeSDR-USB") {
		slot := hotplug.RequestedSlotSpec{
			Attrs: map[string]interface{}{
				"path":      filepath.Clean(di.DeviceName()),
				"sysfspath": filepath.Clean(di.DevicePath()),
				"major":     di.Major(),
				"minor":     di.Minor(),
			},
		}
		return spec.SetSlot(&slot)
	}
	return nil
}

func (iface *limeSdrInterface) BeforePrepareSlot(slot *snap.SlotInfo) error {
	if err := sanitizeSlotReservedForOS(iface, slot); err != nil {
		return err
	}

	var path string
	if err := slot.Attr("path", &path); err != nil {
		return fmt.Errorf("lime-sdr slot must have a path attribute: %s", err)
	}
	// TODO: sysfspath
	return nil
}

func (iface *limeSdrInterface) AppArmorConnectedPlug(spec *apparmor.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
	var path, sysfsPath, major, minor string
	if err := slot.Attr("path", &path); err != nil {
		return err
	}
	if err := slot.Attr("sysfspath", &sysfsPath); err != nil {
		return err
	}
	if err := slot.Attr("major", &major); err != nil {
		return err
	}
	if err := slot.Attr("minor", &minor); err != nil {
		return err
	}

	snippet := strings.Replace(limeSdrConnectedPlugApparmor, "###USB_DEVICE###", path, -1)
	snippet = strings.Replace(snippet, "###SYSFS_PATH###", sysfsPath, -1)
	snippet = strings.Replace(snippet, "###MAJOR###", major, -1)
	snippet = strings.Replace(snippet, "###MINOR###", minor, -1)
	spec.AddSnippet(snippet)
	return nil
}

func (iface *limeSdrInterface) SecCompConnectedPlug(spec *seccomp.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
	spec.AddSnippet(limeSdrConnectedPlugSeccomp)
	return nil
}

func (iface *limeSdrInterface) AutoConnect(*snap.PlugInfo, *snap.SlotInfo) bool {
	// allow what declarations allowed
	return true
}

func init() {
	registerIface(&limeSdrInterface{})
}
