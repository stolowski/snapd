// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
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
	"github.com/snapcore/snapd/interfaces"
)

const upowerControlConnectedPlugAppArmor = `
# Description: Can query UPower as well as suspend and hibernate the system

#include <abstractions/dbus-strict>

# Find all devices monitored by UPower
dbus (send)
    bus=system
    path=/org/freedesktop/UPower
    interface=org.freedesktop.UPower
    member=EnumerateDevices
    peer=(label=unconfined),

# Read all properties from UPower and devices
dbus (send)
    bus=system
    path=/org/freedesktop/UPower{,/devices/**}
    interface=org.freedesktop.DBus.Properties
    member=Get{,All}
    peer=(label=unconfined),

dbus (send)
    bus=system
    path=/org/freedesktop/UPower
    interface=org.freedesktop.UPower
    member=GetCriticalAction
    peer=(label=unconfined),

dbus (send)
    bus=system
    path=/org/freedesktop/UPower/devices/**
    interface=org.freedesktop.UPower.Device
    member=GetHistory
    peer=(label=unconfined),

# Receive property changed events
dbus (receive)
    bus=system
    path=/org/freedesktop/UPower{,/devices/**}
    interface=org.freedesktop.DBus.Properties
    member=PropertiesChanged
    peer=(label=unconfined),

# Receive signals from UPower
dbus (receive)
    bus=system
    path=/org/freedesktop/UPower
	interface=org.freedesktop.UPower
    member="{DeviceAdded,DeviceRemoved,DeviceChanged,Changed,Sleeping,Resuming}"
    peer=(label=unconfined),
`

const upowerControlConnectedPlugSecComp = `
# Description: Can query UPower as well as suspend and hibernate the system

# dbus
connect
getsockname
recvfrom
recvmsg
send
sendto
sendmsg
socket
`

// NewUPowerControlInterface returns a new "upower-observe" interface.
func NewUPowerControlInterface() interfaces.Interface {
	return &commonInterface{
		name: "upower-control",
		connectedPlugAppArmor: upowerControlConnectedPlugAppArmor,
		connectedPlugSecComp:  upowerControlConnectedPlugSecComp,
		reservedForOS:         true,
		autoConnect:           true,
	}
}
