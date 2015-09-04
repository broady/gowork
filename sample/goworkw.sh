#!/bin/bash
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
