// Command octane is the OCTANE conformance test runner CLI.
//
// It provides commands for running .story test suites against a CSMS
// (Charge Station Management System), validating story files, listing
// registered keywords, and managing the content-addressed result cache.
//
// Usage:
//
//	octane [--config path] [--verbose] [--no-cache] [--cache-dir dir] <command>
//
// See "octane help <command>" for details on each subcommand.
package main

func main() {
	Execute()
}
