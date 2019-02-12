// Copyright (c) 2019, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
)

var (
	check = flag.Bool("check", false, "enable extra checks")

	name = flag.String("name", "winup_10_64", "name of the virtual machine")
	cpus = flag.Int("cpus", runtime.NumCPU()/2, "number of cpus to assign")
	mem  = flag.Int("mem", 2048, "memory in megabytes to assign")
)

func logf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func fatalf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}

func main() {
	flag.Parse()

	loadState()
	getDownloads()

	// 10s: import, first boot, core settings; see setup.go
	onState(10, importBox, "imported VM image")
	onState(11, tweakBox, "tweaked VM settings")
	onState(12, firstBoot, "first VM boot")
	onState(13, enableAdmin, "enabled admin login")

	// 20s: uninstall and disable clutter
	onState(20, runDebloater, "ran debloater.ps1")
	onState(21, removeOnedrive, "removed onedrive")

	fmt.Println("all done!")
}

var state, lastState progressState

type progressState struct {
	Readme      string `json:"readme"`
	Step        int    `json:"step"`
	Description string `json:"description"`
}

func (p progressState) SnapName() string {
	return fmt.Sprintf("%d: %s", p.Step, p.Description)
}

func loadState() {
	bs, err := ioutil.ReadFile("progress.json")
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bs, &state); err != nil {
		panic(err)
	}
}

var (
	firstStateFn = true
	justKilled   = false
)

func onState(step int, fn func(), description string) {
	if state.Step >= step {
		// already done.
		lastState.Step = step
		lastState.Description = description
		return
	}
	fmt.Fprintf(os.Stderr, "\n== Step %d: %s ==\n\n", step, description)
	if firstStateFn {
		forceShutdown()
		if lastState.Step > 0 {
			// This step might have failed before; try again from the
			// previous snapshot.
			vbox("snapshot", *name, "restore", lastState.SnapName())
		}
	}

	fn()
	firstStateFn = false

	state.Readme = "This file records the last VM setup step that succeeded."
	state.Step = step
	state.Description = description
	vbox("snapshot", *name, "take", state.SnapName(), "--live")

	bs, err := json.MarshalIndent(&state, "", "\t")
	if err != nil {
		panic(err)
	}
	bs = append(bs, '\n')
	if err := ioutil.WriteFile("progress.json", bs, 0666); err != nil {
		panic(err)
	}
}
