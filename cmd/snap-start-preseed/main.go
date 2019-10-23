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
	"io"
	"os"

	"github.com/jessevdk/go-flags"
)

const (
	shortHelp = "Pre-seed snaps in an Ubuntu chroot"
	longHelp  = `
	Pre-seed snaps in an Ubuntu-based chroot directory
`
)

var (
	osGetuid           = os.Getuid
	Stdout   io.Writer = os.Stdout
	Stderr   io.Writer = os.Stderr

	opts struct{}

	parser *flags.Parser = flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash|flags.PassAfterNonOption)
)

func main() {
	parser.ShortDescription = shortHelp
	parser.LongDescription = longHelp

	if err := run(); err != nil {
		fmt.Fprintf(Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if osGetuid() != 0 {
		return fmt.Errorf("must be run as root")
	}

	rest, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	if len(rest) == 0 {
		return fmt.Errorf("need the chroot path as argument")
	}

	chrootDir := rest[0]

	if err := checkChroot(chrootDir); err != nil {
		return err
	}

	cleanup, err := mountCoreOrSnapdSnap(chrootDir)
	if err != nil {
		return err
	}

	err = runSnapdInChroot(chrootDir)

	cleanup()
	return err
}
