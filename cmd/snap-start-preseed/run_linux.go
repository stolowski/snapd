// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/seed"
	"github.com/snapcore/snapd/timings"
)

var mountPath = "/snapd-prebake"
var syscallMount = syscall.Mount

func checkChroot(prebakeChroot string) error {
	exists, isDir, err := osutil.DirExists(prebakeChroot)
	if err != nil {
		return fmt.Errorf("image-prebaking error: %v", err)
	}
	if !exists || !isDir {
		return fmt.Errorf("image-prebaking chroot directory %s doesn't exist or is not a directory", prebakeChroot)
	}

	// sanity checks of the critical mountpoints inside chroot directory
	for _, p := range []string{"/sys/kernel/security/apparmor", "/proc/self", "/dev/mem"} {
		path := filepath.Join(prebakeChroot, p)
		if exists := osutil.FileExists(path); !exists {
			return fmt.Errorf("image-prebaking chroot directory validation error: %s doesn't exist", path)
		}
	}

	return nil
}

func mountCoreOrSnapdSnap(prebakeChroot string) error {
	seedDir := filepath.Join(dirs.SnapSeedDirUnder(prebakeChroot))

	seed, err := seed.Open(seedDir)
	if err != nil {
		return err
	}

	if err := seed.LoadAssertions(nil, nil); err != nil {
		return err
	}

	tm := timings.New(nil)
	if err := seed.LoadMeta(tm); err != nil {
		return err
	}

	var coreSnapPath string

	ess := seed.EssentialSnaps()
	for _, snap := range ess {
		if snap.SnapName() == "core" {
			coreSnapPath = snap.Path
		}
	}

	// create mountpoint for core/snapd
	where := filepath.Join(prebakeChroot, mountPath)
	if err := os.MkdirAll(where, 0755); err != nil {
		return err
	}

	cmd := exec.Command("/bin/mount", "-t", "squashfs", coreSnapPath, where)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cannot mount %s at %s in pre-bake mode: %v; %s", coreSnapPath, where, err, string(out))
	}

	return fmt.Errorf("!!")
}

// runSnapdInChroot runs snapd in a prebake mode in a chroot dir pointed by prebakeChroot.
// The chroot dir is expected to be set-up and ready to use (all critical system directories mounted).
func runSnapdInChroot(prebakeChroot string) error {
	if err := syscall.Chroot(prebakeChroot); err != nil {
		return fmt.Errorf("image-prebaking chroot error: %v", err)
	}

	if err := os.Setenv("SNAPD_PREBAKE_IMAGE", "1"); err != nil {
		return err
	}

	// exec snapd relative to new chroot, e.g. /snapd-prebake/usr/lib/snapd/snapd
	path := filepath.Join(mountPath, "/usr/lib/snapd/snapd")

	// run snapd in pre-baking mode
	cmd := exec.Command(path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("image-prebaking error: %v\n%s\n", err, string(output))
	}

	return nil
}

func cleanup() {
	cmd := exec.Command("/bin/umount", mountPath)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(Stderr, "%v", err)
	}
	if err := os.Remove(mountPath); err != nil {
		fmt.Fprintf(Stderr, "%v", err)
	}
}
