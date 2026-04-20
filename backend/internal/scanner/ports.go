package scanner

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"sync"
	"time"
)

// topPorts contains a small list of common TCP ports to scan footprint.
var topPorts = []int{
	21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445,
	993, 995, 1433, 1521, 1723, 3306, 3389, 5432, 5900, 6379, 8000, 8080, 8443,
}

// doPortScan performs a concurrent TCP connect scan against the target.
func doPortScan(targetUrl string) ModuleResult {
	parsed, err := url.Parse(targetUrl)
	if err != nil {
		return ModuleResult{Module: "ports", Status: "error", Summary: "Invalid URL for port scanning"}
	}

	host := parsed.Hostname()
	if host == "" {
		host = targetUrl // Fallback in case there was no scheme
	}

	var openPorts []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent dials (goroutines)
	sem := make(chan struct{}, 20)

	for _, port := range topPorts {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore slot

		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore slot

			address := fmt.Sprintf("%s:%d", host, p)
			// TCP connect scan with a short timeout
			conn, err := net.DialTimeout("tcp", address, 1500*time.Millisecond)
			if err == nil {
				conn.Close()
				mu.Lock()
				openPorts = append(openPorts, p)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	sort.Ints(openPorts)

	status := "ok"
	if len(openPorts) > 5 {
		status = "warn" // Exposing too many ports might be a misconfiguration
	} else if len(openPorts) == 0 {
		status = "error" // Maybe host is down or deeply firewalled
	}

	summary := fmt.Sprintf("Found %d open TCP ports out of %d checked", len(openPorts), len(topPorts))

	return ModuleResult{
		Module:  "ports",
		Status:  status,
		Summary: summary,
		Details: map[string]any{
			"host":       host,
			"open_ports": openPorts,
		},
	}
}
