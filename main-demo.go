//go:build demo || test
// +build demo test

package main

import (
	"fmt"
	"os"
)

func main() {
	demoNames := GetAllDemoNames()
	if len(os.Args) < 2 || !contains(demoNames, os.Args[1]) {
		fmt.Printf("Usage: %s <DEMO_NAME> [args...]\n", os.Args[0])
		fmt.Printf("\nAvailable demo names: \n")
		for _, name := range demoNames {
			fmt.Printf("  %s\n", name)
		}
		os.Exit(1)
	}

	RunDemo(os.Args[1], os.Args[2:])
}
