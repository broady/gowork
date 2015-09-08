package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var vendor = flag.Bool("vendorall", true, "don't use the system's GOPATH. work entirely in the gowork workspace.")

const confFile = "go.work.conf"

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
		doInit(flag.Arg(1))
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

func doInit(pkg string) {
	if pkg == "" {
		log.Fatal("pass in a base package to `gowork init`. for example: `gowork init user/mypkg`")
	}

	f, err := os.Create(confFile)
	check(err, "create marker")
	defer f.Close()

	conf := &conf{Pkg: pkg}
	err = json.NewEncoder(f).Encode(conf)
	check(err, "could not write to conf file")
}

func doBuild() {
	workdir, _ := findRoot()

	err := os.Mkdir(filepath.Join(workdir, "bin"), 0755)
	if err != nil && !os.IsExist(err) {
		check(err, "create workdir/bin")
	}

	doGoCmd(append(flag.Args(), "-o", filepath.Join("bin", filepath.Base(workdir)))...)
}

func doGoCmd(args ...string) {
	workdir, pkg := findRoot()
	err := os.Chdir(workdir)
	check(err, "need to call gowork init {root-pkg}")

	// Workspace dir
	tmpwork, err := ioutil.TempDir("", "gowork-w-")
	if tmpwork != "" {
		defer os.RemoveAll(tmpwork)
	}
	check(err, "get tmpwork")

	err = os.MkdirAll(filepath.Join(tmpwork, "src", filepath.Dir(pkg)), 0755)
	check(err, "tmpwork")

	err = os.Symlink(workdir, filepath.Join(tmpwork, "src", pkg))
	check(err, "symlink work")

	// Vendor dir
	tmpvend, err := ioutil.TempDir("", "gowork-v-")
	if tmpvend != "" {
		defer os.RemoveAll(tmpvend)
	}
	check(err, "get tmpvend")

	err = os.MkdirAll(tmpvend, 0755)
	check(err, "mk tmpvend")

	err = os.Symlink(filepath.Join(workdir, "vendor"), filepath.Join(tmpvend, "src"))
	check(err, "symlink tmpvend")

	// Run go command.

	cmd := exec.Command("go", args...)
	cmd.Dir = filepath.Join(tmpwork, "src", pkg)
	if gopath := os.Getenv("GOPATH"); gopath != "" && !*vendor {
		cmd.Env = gopathEnv(tmpvend, gopath, tmpwork)
	} else {
		cmd.Env = gopathEnv(tmpvend, tmpwork)
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

func findRoot() (dir string, pkg string) {
	wd, err := os.Getwd()
	check(err, "getwd")
	root := wd
	for {
		_, err := os.Stat(filepath.Join(wd, confFile))
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

	conf := readConf(root)

	return root, conf.Pkg
}

type conf struct {
	Pkg string
}

func readConf(root string) *conf {
	b, err := ioutil.ReadFile(filepath.Join(root, confFile))
	check(err, "could not read conf file")
	conf := &conf{}

	err = json.Unmarshal(b, conf)
	check(err, "could not decode conf file")

	return conf
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
