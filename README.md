# Create a new project
$ mkdir ~/goworkproj
$ cd ~/goworkproj

# Get the gowork tool (once)
$ go get github.com/broady/gowork
$ gowork init

# (if you don't want gowork on your PATH...)
$ gowork install-wrapper

# No need for a GOPATH!
$ unset GOPATH

$ cat 'package main; import "fmt"; func main() { fmt.Println("hello world") }' > main.go
$ gowork run main.go
# or..
$ ./goworkw.sh run main.go

# Chanage main.go to import some third-party packages. Fetch them:
$ gowork get
# or..
$ ./goworkw.sh get

# See that they ended up vendor/
$ ls vendor
