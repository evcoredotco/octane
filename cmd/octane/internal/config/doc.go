// Package config loads, validates, and resolves the octane CLI
// configuration from three sources in ascending priority order:
//
//  1. Built-in defaults ([Default]).
//  2. An octane.yml file on disk ([Load]).
//  3. Environment variables ([ApplyEnv]).
//  4. Command-line flags ([Resolve]).
//
// Callers construct a [Config] by calling [Load], piping the result
// through [ApplyEnv], and finally through [Resolve] with the parsed
// flag overrides. The resulting [Config] is the authoritative runtime
// configuration for the octane run.
package config
