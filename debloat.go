// Copyright (c) 2019, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func pshellf(format string, a ...interface{}) {
	vbox("guestcontrol", *name, "run",
		"--username", adminUser, "--password", password,
		"--", powerShell, "-c", fmt.Sprintf(format, a...))
}

func runScript(path string, args ...string) {
	base := filepath.Base(path)
	vbox("guestcontrol", *name, "copyto",
		"--username", adminUser, "--password", password,
		"--target-directory", tempDir, path)
	pshellAdmin(tempDir + `\` + base + " " + strings.Join(args, " "))
}

// runDebloater runs an MIT-licensed script to remove bloatware that comes
// installed with Windows 10. It also enhances some privacy settings and
// disables certain background services.
func runDebloater() {
	ensureRunning()
	runScript(debloatScript, "-SysPrep", "-Debloat")
}

// removeOnedrive removes OneDrive, which is a separate script from the
// debloater one.
func removeOnedrive() {
	ensureRunning()
	runScript(onedriveScript)
}

func meteredNet() {
	ensureRunning()
	runScript("scripts/ethernet_metered.ps1")
}

func noBackground() {
	ensureRunning()
	runScript("scripts/no_background.ps1")
}
