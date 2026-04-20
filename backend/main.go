package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"secscan/internal/api"
)

func main() {
	r := gin.Default()

	// CORS — allow Next.js frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Scan routes
	scanGroup := r.Group("/api/scan")
	{
		scanGroup.POST("", api.StartScan)
		scanGroup.GET("/:id", api.GetScanResult)
		scanGroup.GET("/:id/stream", api.StreamScanProgress)
		scanGroup.GET("/:id/report.pdf", api.DownloadPDF)
	}

	log.Println("SecScan backend running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
