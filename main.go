/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"os"

	"github.com/holgerjh/genjsonschema-cli/cmd"
)

func main() {
	err := cmd.RootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
