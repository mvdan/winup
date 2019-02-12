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
	short = flag.Bool("short", false, "skip optional checks")
	kill  = flag.Bool("kill", false, "kill the VM if it is running")

	name = flag.String("name", "winup_10_64", "name of the virtual machine")
	cpus = flag.Int("cpus", runtime.NumCPU()/2, "number of cpus to assign")
	mem  = flag.Int("mem", 4096, "memory in megabytes to assign")
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

	onState(1, importBox, "imported VM image")
	onState(2, tweakBox, "tweaked VM settings")
	onState(3, firstBoot, "first VM boot")
	onState(4, enableAdmin, "enabled admin login")
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

var firstStateFn = true

func onState(step int, fn func(), description string) {
	if state.Step >= step {
		// already done.
		lastState.Step = step
		lastState.Description = description
		return
	}
	fmt.Fprintf(os.Stderr, "\n== Step %d ==\n\n", step)
	if firstStateFn {
		if *kill {
			forceShutdown()
		}
		if lastState.Step > 0 {
			// This step might have failed before; try again from the
			// previous snapshot.
			vbox("snapshot", *name, "restore", lastState.SnapName())
		}
	}
	firstStateFn = false

	fn()

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
