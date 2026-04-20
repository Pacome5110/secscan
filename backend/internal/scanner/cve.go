package scanner

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// doCVE fingerprints the tech stack via headers and simulates a CVE lookup
func doCVE(targetUrl string) ModuleResult {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(targetUrl)
	if err != nil {
		return ModuleResult{Module: "cve", Status: "error", Summary: "Request failed."}
	}
	defer resp.Body.Close()

	server := resp.Header.Get("Server")
	xPowered := resp.Header.Get("X-Powered-By")

	techStack := []string{}
	if server != "" {
		techStack = append(techStack, server)
	}
	if xPowered != "" {
		techStack = append(techStack, xPowered)
	}

	if len(techStack) == 0 {
		return ModuleResult{
			Module:  "cve",
			Status:  "ok",
			Summary: "No explicit technology versions leaked in headers.",
		}
	}

	// Mocking NVD/OSV response logic: In a real app we'd query an API.
	// For this academic project, we just flag if versions are exposed.
	status := "warn"
	summary := fmt.Sprintf("Technology Fingerprinted: %v. Please ensure versions are updated.", techStack)

	return ModuleResult{
		Module:  "cve",
		Status:  status,
		Summary: summary,
		Details: techStack,
	}
}
