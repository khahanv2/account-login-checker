package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khahanv2/account-login-checker/internal/logger"
	"github.com/khahanv2/account-login-checker/internal/models"
	"github.com/khahanv2/account-login-checker/internal/processor"
)

// SetupRouter configures the API routes
func SetupRouter() *gin.Engine {
	router := gin.Default()
	
	// Static files
	router.Static("/static", "./static")
	
	// Templates
	router.LoadHTMLGlob("templates/*")
	
	// Create results directory if not exists
	os.MkdirAll("results", os.ModePerm)
	
	// Routes
	router.GET("/", HomeHandler)
	router.GET("/ws", WebSocketHandler)
	router.POST("/upload", UploadHandler)
	router.GET("/download/:filename", DownloadHandler)
	
	return router
}

// HomeHandler renders the main page
func HomeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Hệ Thống Kiểm Tra Tài Khoản",
	})
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(c *gin.Context) {
	// This is just a placeholder, actual WebSocket handling is done in logger package
	c.Status(http.StatusOK)
}

// UploadHandler processes uploaded Excel files
func UploadHandler(c *gin.Context) {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Không thể nhận file: %s", err.Error()),
		})
		return
	}
	
	// Save the file temporarily
	tempFileName := filepath.Join("temp", fmt.Sprintf("upload_%d.xlsx", time.Now().Unix()))
	
	// Make sure temp directory exists
	os.MkdirAll("temp", os.ModePerm)
	
	if err := c.SaveUploadedFile(file, tempFileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Không thể lưu file: %s", err.Error()),
		})
		return
	}
	
	// Process file in a goroutine
	go func() {
		accounts, err := models.LoadAccountsFromExcel(tempFileName)
		if err != nil {
			logger.LogGeneralJSON(logger.LogLevelError, "Lỗi đọc file Excel", map[string]string{
				"error": err.Error(),
			})
			return
		}
		
		// Start processing accounts
		processor.ProcessAccounts(accounts)
		
		// Clean up temp file
		os.Remove(tempFileName)
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Đã nhận file và bắt đầu xử lý",
		"filename": file.Filename,
	})
}

// DownloadHandler serves result files for download
func DownloadHandler(c *gin.Context) {
	filename := c.Param("filename")
	
	// For security, only allow downloading files from results directory
	if !strings.HasPrefix(filename, "success_") && !strings.HasPrefix(filename, "fail_") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}
	
	// Check if file exists
	filePath := filepath.Join("results", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}
	
	// Set appropriate headers for downloading Excel file
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	
	c.File(filePath)
}