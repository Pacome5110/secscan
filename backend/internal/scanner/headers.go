package scanner

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// doHeadersScan performs an HTTP request and measures security hygiene based on headers.
func doHeadersScan(targetUrl string) ModuleResult {
	if !strings.HasPrefix(targetUrl, "http://") && !strings.HasPrefix(targetUrl, "https://") {
		targetUrl = "http://" + targetUrl
	}

	// Create custom client that ignores cert errors for scanning purposes
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 2 {
				return http.ErrUseLastResponse // Don't redirect forever
			}
			return nil
		},
	}

	// Use HEAD to save bandwidth, but some servers require GET to return all headers
	resp, err := client.Get(targetUrl)
	if err != nil {
		return ModuleResult{
			Module:  "headers",
			Status:  "error",
			Summary: "Failed to connect for headers scan: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	foundHeaders := make(map[string]string)
	missingHeaders := []string{}
	score := 100

	// Helper to grade headers
	check := func(header string, penalty int) {
		val := resp.Header.Get(header)
		if val != "" {
			foundHeaders[header] = val
		} else {
			missingHeaders = append(missingHeaders, header)
			score -= penalty
		}
	}

	// Core Security Headers checks
	check("Content-Security-Policy", 30) // Prevents XSS, most important
	check("X-Frame-Options", 20)         // Prevents Clickjacking
	check("X-Content-Type-Options", 10)  // Prevents MIME-sniffing

	// HSTS is only strictly required on HTTPS
	if strings.HasPrefix(targetUrl, "https") {
		check("Strict-Transport-Security", 20)
	}

	// Calculate Grade
	var grade string
	switch {
	case score == 100:
		grade = "A+"
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 60:
		grade = "C"
	case score >= 40:
		grade = "D"
	default:
		grade = "F"
	}

	// Determine status indicator for frontend
	status := "warn"
	if grade == "A+" || grade == "A" {
		status = "ok"
	} else if grade == "F" || grade == "D" {
		status = "error" // Red indicators
	}

	summary := fmt.Sprintf("Headers Grade: %s. Missing %d key headers.", grade, len(missingHeaders))

	return ModuleResult{
		Module:  "headers",
		Status:  status,
		Summary: summary,
		Details: map[string]any{
			"grade":     grade,
			"score":     score,
			"missing":   missingHeaders,
			"found":     foundHeaders,
			"server":    resp.Header.Get("Server"),
			"x_powered": resp.Header.Get("X-Powered-By"), // Checks framework leaks
		},
	}
}
