package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var vendor = flag.Bool("vendorall", true, "don't use the system's GOPATH. work entirely in the gowork workspace.")

const markerFile = ".gowork"

const wrapper = `#!/bin/bash
set -e

goworkdir="$HOME/.gowork"
if [ ! -d "$goworkdir" ]; then
  mkdir -p $HOME/.gowork
fi

if which gowork; then
  gowork $@
else
  if [ ! -e "$goworkdir/bin/gowork" ]; then
    GOPATH="$goworkdir" go get github.com/broady/gowork
  fi
  "$goworkdir/bin/gowork" $@
fi
`

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "install-wrapper":
		doInstallWrapper()
	case "init":
		doInit()
	case "get":
		err := os.Mkdir("vendor", 0755)
		if err != nil && !os.IsExist(err) {
			check(err, "create vendor")
		}
		doGoCmd("get", "-v", "-d")
	case "run":
		doGoCmd(flag.Args()...)
	case "build":
		doBuild()
	default:
		log.Fatal("usage: `gowork [init|get|build|run|install-wrapper]`")
	}
}

func doInstallWrapper() {
	err := ioutil.WriteFile("goworkw.sh", []byte(wrapper), 0755)
	check(err, "could not write wrapper")
}

func doInit() {
	f, err := os.Create(markerFile)
	check(err, "create marker")
	f.Close()

	err = os.Mkdir("vendor", 0755)
	if err != nil && !os.IsExist(err) {
		check(err, "create vendor")
	}
}

func doBuild() {
	workdir := findRoot()

	err := os.Mkdir(filepath.Join(workdir, "bin"), 0755)
	if err != nil && !os.IsExist(err) {
		check(err, "create workdir/bin")
	}

	doGoCmd(append(flag.Args(), "-o", filepath.Join("bin", filepath.Base(workdir)))...)
}

func doGoCmd(args ...string) {
	workdir := findRoot()
	err := os.Chdir(workdir)
	if err != nil {
		log.Fatalf("need to gowork init? %v", err)
	}

	// Vendor dir
	tmpdir, err := ioutil.TempDir("", "gowork-t-")
	if tmpdir != "" {
		defer os.RemoveAll(tmpdir)
	}
	check(err, "get tempdir")

	err = os.MkdirAll(tmpdir, 0755)
	check(err, "mktempdir")

	err = os.Symlink(filepath.Join(workdir, "vendor"), filepath.Join(tmpdir, "src"))
	check(err, "symlink")

	// Workspace dir
	tmpwork, err := ioutil.TempDir("", "gowork-w-")
	if tmpdir != "" {
		defer os.RemoveAll(tmpwork)
	}
	check(err, "get tmpwork")

	err = os.MkdirAll(tmpwork, 0755)
	check(err, "tmpwork")

	err = os.Symlink(filepath.Join(workdir), filepath.Join(tmpwork, "src"))
	check(err, "symlink")

	// Run go command.

	cmd := exec.Command("go", args...)
	cmd.Dir = filepath.Join(tmpwork, "src")
	if gopath := os.Getenv("GOPATH"); gopath != "" && !*vendor {
		cmd.Env = gopathEnv(tmpdir, gopath, tmpwork)
	} else {
		cmd.Env = gopathEnv(tmpdir, tmpwork)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	check(cmd.Run(), "run go cmd")
}

func gopathEnv(dirs ...string) []string {
	gopath := "GOPATH=" + strings.Join(dirs, string(filepath.ListSeparator))

	env := os.Environ()
	for i, v := range env {
		if strings.HasPrefix(v, "GOPATH=") {
			env[i] = gopath
			return env
		}
	}
	return append(env, gopath)
}

func findRoot() string {
	wd, err := os.Getwd()
	check(err, "getwd")
	root := wd
	for {
		_, err := os.Stat(filepath.Join(wd, markerFile))
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		newRoot := filepath.Dir(root)
		if newRoot == root {
			log.Fatal("not inside a gowork workspace. use gowork init to start one.")
		}
		root = newRoot
	}
	return root
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
