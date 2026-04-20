package api

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	"secscan/internal/scanner"
)

// ScanRequest — POST /api/scan body
type ScanRequest struct {
	URL     string   `json:"url" binding:"required"`
	Modules []string `json:"modules"` // empty = run all
}

// ScanJob — stored in memory (replace with DB for production)
type ScanJob struct {
	ID        string         `json:"id"`
	URL       string         `json:"url"`
	Status    string         `json:"status"` // queued | running | done | error
	CreatedAt time.Time      `json:"created_at"`
	Results   map[string]any `json:"results,omitempty"`
	Progress  []string       `json:"progress,omitempty"`
	mu        sync.Mutex
}

var (
	jobs   = make(map[string]*ScanJob)
	jobsMu sync.RWMutex
)

// StartScan — POST /api/scan
func StartScan(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate & block SSRF
	if err := validateURL(req.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	job := &ScanJob{
		ID:        id,
		URL:       req.URL,
		Status:    "queued",
		CreatedAt: time.Now(),
		Results:   make(map[string]any),
	}

	jobsMu.Lock()
	jobs[id] = job
	jobsMu.Unlock()

	// Run scan asynchronously
	go runScan(job, req.Modules)

	c.JSON(http.StatusAccepted, gin.H{"scan_id": id, "status": "queued"})
}

// GetScanResult — GET /api/scan/:id
func GetScanResult(c *gin.Context) {
	id := c.Param("id")

	jobsMu.RLock()
	job, ok := jobs[id]
	jobsMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	job.mu.Lock()
	defer job.mu.Unlock()
	c.JSON(http.StatusOK, job)
}

// StreamScanProgress — GET /api/scan/:id/stream (SSE)
func StreamScanProgress(c *gin.Context) {
	id := c.Param("id")

	jobsMu.RLock()
	job, ok := jobs[id]
	jobsMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	sent := 0
	for {
		select {
		case <-ticker.C:
			job.mu.Lock()
			progress := job.Progress
			status := job.Status
			job.mu.Unlock()

			// Send new progress lines
			for i := sent; i < len(progress); i++ {
				fmt.Fprintf(c.Writer, "data: %s\n\n", progress[i])
				sent++
			}
			c.Writer.Flush()

			if status == "done" || status == "error" {
				fmt.Fprintf(c.Writer, "data: [DONE] status=%s\n\n", status)
				c.Writer.Flush()
				return
			}

		case <-c.Request.Context().Done():
			return
		}
	}
}

// DownloadPDF generates and streams a PDF report (F09)
func DownloadPDF(c *gin.Context) {
	id := c.Param("id")

	jobsMu.RLock()
	job, exists := jobs[id]
	jobsMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "SecScan Security Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 10, fmt.Sprintf("Target URL: %s", job.URL), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Scan ID: %s", job.ID), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Status: %s", job.Status), "", 1, "L", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Module Results:")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	for mod, res := range job.Results {
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(0, 8, fmt.Sprintf("[%s]", mod), "", 1, "L", false, 0, "")
		
		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(0, 6, fmt.Sprintf("%v", res), "", "L", false)
		pdf.Ln(4)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=SecScan_%s.pdf", id))
	err := pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize PDF"})
	}
}

// runScan — executes all requested scanner modules concurrently
func runScan(job *ScanJob, modules []string) {
	job.mu.Lock()
	job.Status = "running"
	job.mu.Unlock()

	allModules := []string{"ports", "headers", "tls", "fuzz", "xss", "sqli", "cve"}
	if len(modules) > 0 {
		allModules = modules
	}

	var wg sync.WaitGroup
	for _, mod := range allModules {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			job.mu.Lock()
			job.Progress = append(job.Progress, fmt.Sprintf("[%s] starting...", name))
			job.mu.Unlock()

			result := scanner.Run(name, job.URL)

			job.mu.Lock()
			job.Results[name] = result
			job.Progress = append(job.Progress, fmt.Sprintf("[%s] done", name))
			job.mu.Unlock()
		}(mod)
	}

	wg.Wait()

	job.mu.Lock()
	job.Status = "done"
	job.mu.Unlock()
}

// validateURL — blocks SSRF via private IP ranges
func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http/https allowed")
	}
	host := parsed.Hostname()
	if isPrivateHost(host) {
		return fmt.Errorf("private/internal hosts are not allowed (SSRF protection)")
	}
	return nil
}

func isPrivateHost(host string) bool {
	privateHosts := []string{
		"localhost", "127.", "10.", "192.168.", "172.16.", "172.17.",
		"172.18.", "172.19.", "172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.", "172.28.", "172.29.",
		"172.30.", "172.31.", "169.254.", "::1", "0.0.0.0",
	}
	for _, p := range privateHosts {
		if host == p || len(host) > len(p) && host[:len(p)] == p {
			return true
		}
	}
	return false
}
