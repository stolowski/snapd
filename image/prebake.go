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

package image

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func createMounts(rootDir string) (func(), error) {
	mountPoints := []string{}
	dirs := []string{}
	restore := func() {
		for _, m := range mountPoints {
			fmt.Printf("unmounting: %s\n", m)
			if err := syscall.Unmount(m, 0); err != nil {
				fmt.Printf("error: %v\n", err)
			}
		}
		for _, d := range dirs {
			fmt.Printf("removing: %s\n", d)
		}
	}

	for _, d := range []string{"run", "proc", "bin", "sbin"} {
		if err := os.Mkdir(filepath.Join(rootDir, d), 0755); err != nil {
			return nil, err
		}
		dirs = append(dirs)
	}

	bindFlags := syscall.MS_BIND
	for _, m := range []string{"proc", "bin", "sbin"} {
		imageTargetDir := filepath.Join(rootDir, m)
		if err := syscall.Mount(filepath.Join("/", m), imageTargetDir, "proc", uintptr(bindFlags), ""); err != nil {
			return nil, err
		}
		mountPoints = append(mountPoints, imageTargetDir)
	}

	return restore, nil
}

func Prebake(opts *Options) error {
	path, err := filepath.Abs(opts.RootDir)
	if err != nil {
		return err
	}
	os.Setenv("SNAPD_PREBAKE_IMAGE", path)
	os.Setenv("SNAP_REEXEC", "0")

	_, err = createMounts(path)
	if err != nil {
		return fmt.Errorf("failed to create mounts: %v\n", err)
	}
	//defer restoreMounts()

	cmd := exec.Command("/home/pawel/go/src/github.com/snapcore/snapd/cmd/snapd/snapd")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run snapd:\n%s\n%v", string(out), err)
	}

	return nil
}
