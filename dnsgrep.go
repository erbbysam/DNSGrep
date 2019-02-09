// a lightweight utility to scan a sorted file for a substring at the start of each line

package main

import (
	. "dnsgrep/DNSBinarySearch"

	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

// command line parsing
type Options struct {
	Input   string `short:"i" long:"input" description:"A hostname to search" required:"true"`
	DNSFile string `short:"f" long:"file" description:"A large file containing sorted, reversed domain names" required:"true"`
}

// command line parsing
var options Options
var parser = flags.NewParser(&options, flags.Default)

func main() {

	// command line parsing
	_, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	// increase our limits x10 as we're running this locally
	var limits = Limits{
		MaxScan:        1000,    // 100MB
		MaxOutputLines: 1000000, // 1,000,000 lines
	}

	// main.go is really just a wrapper around this function
	output, err := DNSBinarySearch(options.DNSFile, options.Input, limits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
	} else {
		for _, result := range output {
			fmt.Printf("%s\n", result)
		}
	}

}
