// Copyright (c) 2019, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import "fmt"

func pshellf(format string, a ...interface{}) {
	vbox("guestcontrol", *name, "run",
		"--username", adminUser, "--password", password,
		"--", powerShell, "-c", fmt.Sprintf(format, a...))
}

// runDebloater runs an MIT-licensed script to remove bloatware that comes
// installed with Windows 10. It also enhances some privacy settings and
// disables certain background services.
func runDebloater() {
	ensureRunning()
	vbox("guestcontrol", *name, "copyto", "--username", adminUser, "--password", password,
		"--target-directory", tempDir, debloatScript)
	pshellAdmin(tempDir + `\debloater.ps1 -SysPrep -Debloat`)
}

// removeOnedrive removes OneDrive, which is a separate script from the
// debloater one.
func removeOnedrive() {
	ensureRunning()
	vbox("guestcontrol", *name, "copyto", "--username", adminUser, "--password", password,
		"--target-directory", tempDir, onedriveScript)
	pshellAdmin(tempDir + `\remove-onedrive.ps1`)
}
