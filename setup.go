// Copyright (c) 2019, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	regUser   = "IEUser"
	adminUser = "Administrator"
	password  = "Passw0rd!"

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

// waitStdout will keep running a command every few seconds until its stdout
// matches a substring.
func waitStdout(substr, name string, args ...string) {
	var out []byte
	var err error
	for i := 0; i < 120; i++ {
		time.Sleep(2 * time.Second)
		out, err = command(name, args...).Output()
		if strings.Contains(string(out), substr) {
			return
		}
	}
	fatalf("timed out waiting for a stdout match: %v\n%s", err, out)
}

func boot() {
	vbox("startvm", *name)
	// wait for the login to finish
	vbox("guestproperty", "wait",
		*name, "/VirtualBox/GuestInfo/OS/LoggedInUsers")
	waitStdout("Value: 1", "vboxmanage", "-q", "guestproperty", "get",
		*name, "/VirtualBox/GuestInfo/OS/LoggedInUsers")

	// waiting for explore.exe doesn't really make us wait longer.
	// TODO: any way to see if we reached the desktop? is that helpful?
	//waitStdout("ProcessName", "vboxmanage", "-q", "guestcontrol",
	//        *name, "run", "--username", "IEUser", "--password", "Passw0rd!",
	//        "--", powerShell, "-c", `Get-Process *explorer*`)
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
	if strings.Contains(string(out), "is not currently running") {
		return
	}
	if err != nil {
		panic(err)
	}
	// apparently virtualbox doesn't release the VM lock instantly
	time.Sleep(time.Second)
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

func enableAdmin() {
	boot() // in case it's not yet running
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// keep pressing ALT+Y until we've run the cmd as admin
		for {
			select {
			case <-ctx.Done():
				return
			default:
				inputCodes(alt(ascii("y")))
				time.Sleep(time.Second)
			}
		}
	}()
	run("vboxmanage", "-q", "guestcontrol", *name, "run",
		"--username", "IEUser", "--password", "Passw0rd!",
		"--", powerShell, "-c", `Start-Process -verb runAs powershell.exe -argumentlist "net user Administrator /active:yes"`)
	cancel()
}
