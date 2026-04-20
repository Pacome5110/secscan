// Package scanner provides the registry and stub implementations for all
// SecScan scanner modules. Each module is invoked by name via Run().
package scanner

import "fmt"

// ModuleResult — common result envelope
type ModuleResult struct {
	Module  string `json:"module"`
	Status  string `json:"status"` // ok | warn | error | stub
	Summary string `json:"summary"`
	Details any    `json:"details,omitempty"`
}

// Run dispatches to the correct scanner module by name.
// Modules not yet implemented return a stub result.
func Run(name, target string) ModuleResult {
	switch name {
	case "ports":
		return runPorts(target)
	case "headers":
		return runHeaders(target)
	case "tls":
		return runTLS(target)
	case "fuzz":
		return runFuzz(target)
	case "xss":
		return runXSS(target)
	case "sqli":
		return runSQLi(target)
	case "cve":
		return runCVE(target)
	default:
		return ModuleResult{
			Module:  name,
			Status:  "error",
			Summary: fmt.Sprintf("unknown module: %s", name),
		}
	}
}

// ---------------------------------------------------------------------------
// Module stubs — will be replaced with real implementations in F03–F07
// ---------------------------------------------------------------------------

func runPorts(target string) ModuleResult {
	return doPortScan(target)
}

func runHeaders(target string) ModuleResult {
	return doHeadersScan(target)
}

func runTLS(target string) ModuleResult {
	return doTLSScan(target)
}

func runFuzz(target string) ModuleResult {
	return doFuzz(target)
}

func runXSS(target string) ModuleResult {
	return doXSS(target)
}

func runSQLi(target string) ModuleResult {
	return doSQLi(target)
}

func runCVE(target string) ModuleResult {
	return doCVE(target)
}
