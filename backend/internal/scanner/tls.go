package scanner

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// doTLSScan analyzes the TLS certificate and configuration of the target.
func doTLSScan(targetUrl string) ModuleResult {
	if !strings.HasPrefix(targetUrl, "https://") {
		return ModuleResult{
			Module:  "tls",
			Status:  "warn",
			Summary: "Target is not using HTTPS. TLS scan skipped.",
		}
	}

	parsed, err := url.Parse(targetUrl)
	if err != nil {
		return ModuleResult{Module: "tls", Status: "error", Summary: "Invalid URL"}
	}

	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "443"
	}
	address := fmt.Sprintf("%s:%s", host, port)

	conf := &tls.Config{
		InsecureSkipVerify: true, // We want to inspect the cert even if it's invalid
	}

	client, err := tls.DialWithDialer(&netDialer(), "tcp", address, conf)
	if err != nil {
		return ModuleResult{
			Module:  "tls",
			Status:  "error",
			Summary: "Failed to perform TLS handshake: " + err.Error(),
		}
	}
	defer client.Close()

	state := client.ConnectionState()
	certs := state.PeerCertificates

	if len(certs) == 0 {
		return ModuleResult{Module: "tls", Status: "error", Summary: "No certificates presented"}
	}

	cert := certs[0]
	expiry := cert.NotAfter
	daysUntilExpiry := int(time.Until(expiry).Hours() / 24)

	status := "ok"
	summary := fmt.Sprintf("Valid TLS Certificate. Expires in %d days.", daysUntilExpiry)

	if daysUntilExpiry < 30 {
		status = "warn"
		summary = fmt.Sprintf("Certificate expiring soon! (%d days)", daysUntilExpiry)
	}
	if daysUntilExpiry < 0 {
		status = "error"
		summary = "Certificate is EXPIRED!"
	}

	// Figure out TLS version
	tlsVersion := "Unknown"
	switch state.Version {
	case tls.VersionTLS10:
		tlsVersion = "TLS 1.0 (Deprecated!)"
		status = "error"
	case tls.VersionTLS11:
		tlsVersion = "TLS 1.1 (Deprecated!)"
		status = "error"
	case tls.VersionTLS12:
		tlsVersion = "TLS 1.2"
	case tls.VersionTLS13:
		tlsVersion = "TLS 1.3"
	}

	return ModuleResult{
		Module:  "tls",
		Status:  status,
		Summary: summary,
		Details: map[string]any{
			"issuer_cn":       cert.Issuer.CommonName,
			"subject_cn":      cert.Subject.CommonName,
			"expires_at":      expiry.Format(time.RFC3339),
			"days_left":       daysUntilExpiry,
			"tls_version":     tlsVersion,
			"cipher_suite_id": state.CipherSuite,
		},
	}
}

// helper for timeout
func netDialer() *net.Dialer {
	return &net.Dialer{Timeout: 5 * time.Second}
}
