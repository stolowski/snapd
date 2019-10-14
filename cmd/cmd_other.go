// -*- Mode: Go; indent-tabs-mode: t -*-
// +build !linux

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

package cmd

import (
	"errors"
)

// ExecInSnapdOrCoreSnap makes sure you're executing the binary that ships in
// the snapd/core snap.
// On this OS this is a stub.
func ExecInSnapdOrCoreSnap() {
	return
}

// PrebakeChroot runs snapd in a chroot dir pointed by SNAPD_PREBAKE_IMAGE environment variable
func PrebakeChroot() error {
	return errors.New("image-prebaking not available for systems other than linux")
}

// InternalToolPath returns the path of an internal snapd tool. The tool
// *must* be located inside the same tree as the current binary.
//
// On this OS this is a stub and always returns an error.
func InternalToolPath(tool string) (string, error) {
	return "", errors.New("unsupported on non-Linux systems")
}
