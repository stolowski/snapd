// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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

package release

import (
	"bufio"
	"os"
	"strings"
	"unicode"

	"github.com/snapcore/snapd/apparmor"
)

// Series holds the Ubuntu Core series for snapd to use.
var Series = "16"

// OS contains information about the system extracted from /etc/os-release.
type OS struct {
	ID        string `json:"id"`
	VersionID string `json:"version-id,omitempty"`
}

var (
	apparmorFeaturesSysPath  = "/sys/kernel/security/apparmor/features"
	requiredApparmorFeatures = []string{
		"caps",
		"dbus",
		"domain",
		"file",
		"mount",
		"namespaces",
		"network",
		"ptrace",
		"signal",
	}
)

// ForceDevMode returns true if the distribution doesn't implement required
// security features for confinement and devmode is forced.
func (o *OS) ForceDevMode() bool {
	return AppArmorSupportLevel() != apparmor.FullSupport
}

// AppArmorSupportLevel quantifies how well apparmor is supported on the current kernel.
func AppArmorSupportLevel() apparmor.SupportLevel {
	level, _ := aa.SupportLevel()
	return level
}

// AppArmorSupportSummary describes how well apparmor is supported on the current kernel.
func AppArmorSupportSummary() string {
	_, summary := aa.SupportLevel()
	return summary
}

// AppArmorSupports returns true if the given apparmor feature is supported.
func AppArmorSupports(feature string) bool {
	return aa.SupportsFeature(feature)
}

var (
	osReleasePath         = "/etc/os-release"
	fallbackOsReleasePath = "/usr/lib/os-release"
)

// readOSRelease returns the os-release information of the current system.
func readOSRelease() OS {
	// TODO: separate this out into its own thing maybe (if made more general)
	osRelease := OS{
		// from os-release(5): If not set, defaults to "ID=linux".
		ID: "linux",
	}

	f, err := os.Open(osReleasePath)
	if err != nil {
		// this fallback is as per os-release(5)
		f, err = os.Open(fallbackOsReleasePath)
		if err != nil {
			return osRelease
		}
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ws := strings.SplitN(scanner.Text(), "=", 2)
		if len(ws) < 2 {
			continue
		}

		k := strings.TrimSpace(ws[0])
		v := strings.TrimFunc(ws[1], func(r rune) bool { return r == '"' || r == '\'' || unicode.IsSpace(r) })
		// XXX: should also unquote things as per os-release(5) but not needed yet in practice
		switch k {
		case "ID":
			// ID should be “A lower-case string (no spaces or
			// other characters outside of 0–9, a–z, ".", "_" and
			// "-") identifying the operating system, excluding any
			// version information and suitable for processing by
			// scripts or usage in generated filenames.”
			//
			// So we mangle it a little bit to account for people
			// not being too good at reading comprehension.
			// Works around e.g. lp:1602317
			osRelease.ID = strings.Fields(strings.ToLower(v))[0]
		case "VERSION_ID":
			osRelease.VersionID = v
		}
	}

	return osRelease
}

// OnClassic states whether the process is running inside a
// classic Ubuntu system or a native Ubuntu Core image.
var OnClassic bool

// ReleaseInfo contains data loaded from /etc/os-release on startup.
var ReleaseInfo OS

// aa contains information about available apparmor feature set.
var aa *apparmor.KernelSupport

func init() {
	ReleaseInfo = readOSRelease()

	OnClassic = (ReleaseInfo.ID != "ubuntu-core")

	aa = apparmor.ProbeKernel()
}

// MockOnClassic forces the process to appear inside a classic
// Ubuntu system or a native image for testing purposes.
func MockOnClassic(onClassic bool) (restore func()) {
	old := OnClassic
	OnClassic = onClassic
	return func() { OnClassic = old }
}

// MockReleaseInfo fakes a given information to appear in ReleaseInfo,
// as if it was read /etc/os-release on startup.
func MockReleaseInfo(osRelease *OS) (restore func()) {
	old := ReleaseInfo
	ReleaseInfo = *osRelease
	return func() { ReleaseInfo = old }
}

// MockForcedDevmode fake the system to believe its in a distro
// that is in ForcedDevmode
func MockForcedDevmode(isDevmode bool) (restore func()) {
	level := apparmor.FullSupport
	if isDevmode {
		level = apparmor.NoSupport
	}
	return MockAppArmorSupportLevel(level)
}

// MockAppArmorSupportLevel makes the system believe it has certain level of apparmor support.
func MockAppArmorSupportLevel(level apparmor.SupportLevel) (restore func()) {
	r := apparmor.MockSupportLevel(level)
	old := aa
	aa = apparmor.ProbeKernel()
	return func() {
		r()
		aa = old
	}
}
