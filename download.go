// Copyright (c) 2019, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/mitchellh/ioprogress"
)

const (
	win10Zip = "cache/win10.zip"
	win10Ova = "cache/win10.ova"

	guestIso       = "cache/guest6.0.4.iso"
	debloatScript  = "cache/debloater.ps1"
	onedriveScript = "cache/remove-onedrive.ps1"

	goInst  = "cache/go1.11.5.windows-amd64.msi"
	gitInst = "cache/git-2.20.1-amd64.exe"
)

type download struct {
	name, url string
	sha256sum string
}

var downloads = []download{
	{
		win10Zip,
		"https://az792536.vo.msecnd.net/vms/VMBuild_20180425/VirtualBox/MSEdge/MSEdge.Win10.VirtualBox.zip",
		"36c13632cc9769373262bf041f2a81cc2cbbb0417ebfd965a2bc5a3c7f4e38e7",
	},
	{
		debloatScript,
		"https://raw.githubusercontent.com/Sycnex/Windows10Debloater/65a651f262d67fb69080ff1f26c698231db383ff/Windows10SysPrepDebloater.ps1",
		"fd6fbe791c2762a050ec263e69ccd3450332d50bcee5c182cbe3ddfd30449d26",
	},
	{
		onedriveScript,
		"https://raw.githubusercontent.com/Sycnex/Windows10Debloater/65a651f262d67fb69080ff1f26c698231db383ff/Individual%20Scripts/Uninstall%20OneDrive",
		"b953da06b98d28e173d4c948a8b0efcc47c709df86204b1f897b86257dc97960",
	},
	/*
		{
			guestIso,
			"https://download.virtualbox.org/virtualbox/6.0.4/VBoxGuestAdditions_6.0.4.iso",
			"749b0c76aa6b588e3310d718fc90ea472fdc0b7c8953f7419c20be7e7fa6584a",
		},
		{
			goInst,
			"https://dl.google.com/go/go1.11.5.windows-amd64.msi",
			"01058e46f14f16d2817c762963dbd787b8326c421573bac1624cf7afbbbd499b",
		},
		{
			gitInst,
			"https://github.com/git-for-windows/git/releases/download/v2.20.1.windows.1/Git-2.20.1-64-bit.exe",
			"0dce453188d4aed938e3fd1919393a3600dd3dfe100f3fc92f54f80e372e031f",
		},
	*/
}

func getDownloads() {
	if err := os.MkdirAll("cache", 0777); err != nil {
		panic(err)
	}
	var todo []download
	var wg sync.WaitGroup
	for _, dw := range downloads {
		if !fileExists(dw.name) {
			todo = append(todo, dw)
			continue
		}
		if !*check {
			continue
		}
		wg.Add(1)
		go func(dw download) {
			sum := hashFile(dw.name)
			if sum != dw.sha256sum {
				fatalf("%s checksum mismatch!", dw.name)
			}
			logf("%s already downloaded", dw.name)
			wg.Done()
		}(dw)
	}
	wg.Wait()

	for _, dw := range todo {
		resp := httpGet(dw.url)
		bodySize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		logf("Downloading %s", dw.url)
		r := &ioprogress.Reader{
			Reader:   resp.Body,
			Size:     bodySize,
			DrawFunc: ioprogress.DrawTerminalf(os.Stderr, ioprogress.DrawTextFormatBytes),
		}
		f := createFile(dw.name)
		if _, err := io.Copy(f, r); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
		wg.Add(1)
		go func(dw download) {
			sum := hashFile(dw.name)
			if sum != dw.sha256sum {
				fatalf("%s checksum mismatch!", dw.name)
			}
			wg.Done()
		}(dw)
	}
	wg.Wait()

	extractZip(win10Zip, win10Ova)
}

func extractZip(from, to string) {
	if fileExists(to) {
		return
	}
	zipr, err := zip.OpenReader(from)
	if err != nil {
		panic(err)
	}
	defer zipr.Close()
	if len(zipr.File) != 1 {
		fatalf("expected one file inside zip, got %d", len(zipr.File))
	}
	zipf, err := zipr.File[0].Open()
	if err != nil {
		panic(err)
	}
	f := createFile(to)
	if _, err := io.Copy(f, zipf); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func createFile(path string) *os.File {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	return f
}

func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func httpGet(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode >= 400 {
		fatalf("GET %s: %d", url, resp.StatusCode)
	}
	return resp
}
