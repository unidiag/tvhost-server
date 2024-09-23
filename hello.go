package main

import (
	"bufio"
	"fmt"
	"os"
)

func hello() {
	fmt.Println("****************************************************************")
	fmt.Println("*            Television HOST - https://tvhost.cc               *")
	fmt.Println("*            Copyright Vitali @unidiag Tumasheuski             *")
	fmt.Println("*               This program is server for TVHOST              *")
	fmt.Println("*            Ver: " + _version_[2] + ". Build: " + _version_[0] + " " + _version_[1] + "            *")
	fmt.Println("****************************************************************")
	fmt.Println("     Just run the binary file in Linux (without parameters),")
	fmt.Println("              and we will do the rest ourselves..")
	fmt.Println("\n                    [Press ENTER to exit]")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	os.Exit(0)
}
