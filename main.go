package main

import (
	"github.com/tkancf/mdtask/cmd/mdtask"
)

// Version information set by build flags
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	mdtask.SetVersionInfo(version, commit, buildTime)
	mdtask.Execute()
}