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
	// TODO F04: HTTP security header checks + grade
	return ModuleResult{Module: "headers", Status: "stub", Summary: "Headers checker not yet implemented (F04)"}
}

func runTLS(target string) ModuleResult {
	// TODO F05: TLS version + cipher + cert auditor
	return ModuleResult{Module: "tls", Status: "stub", Summary: "TLS auditor not yet implemented (F05)"}
}

func runFuzz(target string) ModuleResult {
	// TODO F06: Directory fuzzer + wordlist
	return ModuleResult{Module: "fuzz", Status: "stub", Summary: "Dir fuzzer not yet implemented (F06)"}
}

func runXSS(target string) ModuleResult {
	// TODO F06: XSS payload injection + reflection check
	return ModuleResult{Module: "xss", Status: "stub", Summary: "XSS scanner not yet implemented (F06)"}
}

func runSQLi(target string) ModuleResult {
	// TODO F06: Error/boolean/time-based SQLi detection
	return ModuleResult{Module: "sqli", Status: "stub", Summary: "SQLi scanner not yet implemented (F06)"}
}

func runCVE(target string) ModuleResult {
	// TODO F07: Tech detection + NVD/OSV API lookup
	return ModuleResult{Module: "cve", Status: "stub", Summary: "CVE scanner not yet implemented (F07)"}
}
