package scanner

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var commonPaths = []string{
	"/.env", "/admin", "/.git/config", "/wp-admin", "/phpmyadmin",
	"/backup.zip", "/config.yml", "/api/v1/users", "/server-status",
}

// doFuzz performs directory/file fuzzing
func doFuzz(targetUrl string) ModuleResult {
	if !strings.HasSuffix(targetUrl, "/") {
		targetUrl += "/"
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects for fuzzing
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	foundPaths := []string{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, path := range commonPaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			reqUrl := targetUrl + strings.TrimPrefix(p, "/")
			resp, err := client.Get(reqUrl)
			if err == nil {
				// We consider 200 OK or 403 Forbidden as "discovered" paths
				if resp.StatusCode == 200 || resp.StatusCode == 403 {
					mu.Lock()
					foundPaths = append(foundPaths, fmt.Sprintf("%s (%d)", p, resp.StatusCode))
					mu.Unlock()
				}
				resp.Body.Close()
			}
		}(path)
	}

	wg.Wait()

	status := "ok"
	if len(foundPaths) > 0 {
		status = "warn"
	}

	return ModuleResult{
		Module:  "fuzz",
		Status:  status,
		Summary: fmt.Sprintf("Fuzzer found %d sensitive/hidden paths out of %d checked.", len(foundPaths), len(commonPaths)),
		Details: foundPaths,
	}
}

// doXSS attempts payload reflection
func doXSS(targetUrl string) ModuleResult {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	payload := "<script>alert('secscan_xss')</script>"

	// Try appending as a query param
	reqUrl := targetUrl
	if strings.Contains(reqUrl, "?") {
		reqUrl += "&q=" + url.QueryEscape(payload)
	} else {
		reqUrl += "?search=" + url.QueryEscape(payload)
	}

	resp, err := client.Get(reqUrl)
	if err != nil {
		return ModuleResult{Module: "xss", Status: "error", Summary: "Request failed."}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	// If the exact pure HTML payload reflects, it's highly vulnerable
	status := "ok"
	summary := "No obvious reflected XSS detected."
	if strings.Contains(bodyStr, payload) {
		status = "error"
		summary = "Reflected XSS Vulnerability DETECTED!"
	}

	return ModuleResult{
		Module:  "xss",
		Status:  status,
		Summary: summary,
	}
}

// doSQLi attempts basic error-based SQL Injection detection
func doSQLi(targetUrl string) ModuleResult {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	payload := "' OR 1=1 --"
	reqUrl := targetUrl
	if strings.Contains(reqUrl, "?") {
		reqUrl += "&id=" + url.QueryEscape(payload)
	} else {
		reqUrl += "?id=" + url.QueryEscape(payload)
	}

	resp, err := client.Get(reqUrl)
	if err != nil {
		return ModuleResult{Module: "sqli", Status: "error", Summary: "Request failed."}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	status := "ok"
	summary := "No obvious SQLi errors detected."

	// Simple heuristic for generic SQL errors leaking
	errorSignatures := []string{
		"SQL syntax",
		"mysql_fetch_array",
		"ORA-01756",
		"PostgreSQL query failed",
		"SQLite/JDBCDriver",
	}

	for _, sig := range errorSignatures {
		if strings.Contains(bodyStr, sig) {
			status = "error"
			summary = "Potential SQL Injection detected (Database error leaked)!"
			break
		}
	}

	return ModuleResult{
		Module:  "sqli",
		Status:  status,
		Summary: summary,
	}
}
