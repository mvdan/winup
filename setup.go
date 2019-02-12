// Copyright (c) 2019, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	regUser   = "ieuser"
	adminUser = "administrator"
	password  = "Passw0rd!"

	tempDir    = `C:\Users\administrator`
	powerShell = `C:\Windows\System32\WindowsPowershell\v1.0\powershell.exe`
)

func command(name string, args ...string) *exec.Cmd {
	fmt.Fprintf(os.Stderr, "$ %s %s\n", name, strings.Join(args, " "))
	return exec.Command(name, args...)
}

func run(name string, args ...string) {
	cmd := command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func vbox(args ...string) {
	opts := []string{"-q"}
	run("vboxmanage", append(opts, args...)...)
}

// importBox imports the OVA file into a VM with some basic settings.
func importBox() {
	vbox("import", "--vsys", "0",
		"--ostype", "Windows10_64",
		"--vmname", *name,
		"--cpus", strconv.Itoa(*cpus),
		"--memory", strconv.Itoa(*mem),
		win10Ova)
}

// tweakBox applies some extra settings that we couldn't do in importBox.
func tweakBox() {
	vbox("modifyvm", *name,
		"--cableconnected1", "off",
	)
}

func matchStdout(substr, name string, args ...string) bool {
	out, err := command(name, args...).Output()
	return err == nil && strings.Contains(string(out), substr)
}

// waitStdout will keep running a command every few seconds until its stdout
// matches a substring. If substr begins with "! ", it is a negative match.
func waitStdout(substr, name string, args ...string) {
	// usually it's not ready right away, but make the first sleep faster
	time.Sleep(500 * time.Millisecond)
	var out []byte
	var err error
	negative := strings.HasPrefix(substr, "! ")
	if negative {
		substr = substr[2:]
	}
	for i := 0; i < 120; i++ {
		out, err = command(name, args...).Output()
		if strings.Contains(string(out), substr) == !negative {
			return
		}
		time.Sleep(2 * time.Second)
	}
	fatalf("timed out waiting for a stdout match: %v\n%s", err, out)
}

func boot() {
	vbox("startvm", *name)
	// wait for the login to finish; once for the guest to respond, a second
	// time for the value to change on the login
	for i := 0; i < 2; i++ {
		vbox("guestproperty", "wait", *name, "/VirtualBox/GuestInfo/OS/LoggedInUsers")
	}
	waitStdout("Value: 1", "vboxmanage", "-q", "guestproperty", "get",
		*name, "/VirtualBox/GuestInfo/OS/LoggedInUsers")

	// waiting for explore.exe doesn't really make us wait longer.
	// TODO: any way to see if we reached the desktop? is that helpful?
	//waitStdout("ProcessName", "vboxmanage", "-q", "guestcontrol",
	//        *name, "run", "--username", regUser, "--password", password,
	//        "--", powerShell, "-c", `get-process *explorer*`)
}

func ensureRunning() {
	if matchStdout(`VMState="saved"`, "vboxmanage", "-q",
		"showvminfo", "--machinereadable", *name) {
		// already running, but saved
		boot()
		return
	}
	if matchStdout("Value: 1", "vboxmanage", "-q", "guestproperty", "get",
		*name, "/VirtualBox/GuestInfo/OS/LoggedInUsers") {
		// already running and logged in
		return
	}
	if firstStateFn || justKilled {
		boot()
	}
}

// shutdown is a proper shutdown action, letting the guest OS halt normally.
func shutdown() {
	vbox("controlvm", *name, "acpipowerbutton")
	// wait for the shutdown to finish
	waitStdout(`VMState="poweroff"`, "vboxmanage", "-q",
		"showvminfo", "--machinereadable", *name)
	// sometimes the VM lock isn't released immediately; sleep?
	forceShutdown()
}

// forceShutdown shuts off the VM immediately, which is faster but not as safe
// as shutdown.
func forceShutdown() {
	cmd := command("vboxmanage", "-q", "controlvm", *name, "poweroff")
	out, err := cmd.CombinedOutput()
	switch {
	case strings.Contains(string(out), "is not currently running"):
		return
	case strings.Contains(string(out), "not find a registered machine"):
		return
	}
	if err != nil {
		panic(err)
	}
	// apparently virtualbox doesn't release the VM lock instantly
	time.Sleep(time.Second)
	justKilled = true
}

// firstBoot ensures that the first VM boot succeeds, which can take a while.
// Important that this happens without a network connection, to keep Windows
// from installing random apps.
func firstBoot() {
	boot()

	// the first boot can be flaky, so do a power cycle
	shutdown()
	boot()
}

// inputCodes sends a number of keycode sequences to the VM.
func inputCodes(seqs ...[]uint8) {
	args := []string{"controlvm", *name, "keyboardputscancode"}
	args = append(args, codes(seqs...)...)
	vbox(args...)
}

func pshellAdmin(src string) {
	go func() {
		// press ALT+Y once we see the consent window process
		waitStdout("ProcessName", "vboxmanage", "-q", "guestcontrol",
			*name, "run", "--username", adminUser, "--password", password,
			"--", powerShell, "-c", `get-process *consent*`)
		inputCodes(alt(ascii("y")))
	}()
	// we need to run powershell as a regular user first, since we may not
	// have enabled admin login yet, and because some commands require a
	// real terminal window.
	vbox("guestcontrol", *name, "run", "--username", regUser, "--password", password,
		"--", powerShell, "-c", fmt.Sprintf(`start-process -wait -verb runas powershell -argumentlist "%s"`, src))

	// sometimes "start-process -wait" will return before the powershell
	// window has finished running; double-check via get-process
	waitStdout(`! ProcessName`, "vboxmanage", "-q", "guestcontrol", *name,
		"run", "--username", regUser, "--password", password,
		"--", powerShell, "-c",
		`get-process | where-object {$_.MainWindowTitle -eq "Administrator: Windows PowerShell"}`)

}

func enableAdmin() {
	ensureRunning()
	pshellAdmin("net user administrator /active:yes")
}
