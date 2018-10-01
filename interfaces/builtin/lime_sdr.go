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
	"github.com/snapcore/snapd/snap"
)

const limeSdrSummary = `allows accessing a specific serial port`

const limeSdrBaseDeclarationSlots = `
  lime-sdr:
    allow-installation:
      slot-snap-type:
        - core
    deny-auto-connection: true
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

func (iface *limeSdrInterface) HotplugDeviceKey(di *hotplug.HotplugDeviceInfo) (string, error) {
	if di.Subsystem() != "usb" {
		return "", nil
	}

	if model, ok := di.Attribute("ID_MODEL_FROM_DATABASE"); !ok || !strings.Contains(model, "LimeSDR") {
		return "", nil
	}

	// XXX: for some reason "remove" udev event for LimeSDR contains only a small subset of attributes reported by "add",
	// therefore device key can only use PRODUCT attribute.
	product, _ := di.Attribute("PRODUCT")
	return product, nil
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
				"path": filepath.Clean(di.DeviceName()),
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
	return nil
}

func (iface *limeSdrInterface) AppArmorConnectedPlug(spec *apparmor.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
	var path string
	if err := slot.Attr("path", &path); err != nil {
		return err
	}
	spec.AddSnippet(fmt.Sprintf("%s rw,", path))
	return nil
}

func (iface *limeSdrInterface) AutoConnect(*snap.PlugInfo, *snap.SlotInfo) bool {
	// allow what declarations allowed
	return true
}
