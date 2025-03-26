package processor

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/khahanv2/account-login-checker/internal/logger"
	"github.com/khahanv2/account-login-checker/internal/models"
)

// ProcessorConfig holds configuration for the processor
type ProcessorConfig struct {
	MaxConcurrent int
	DelayBetween  time.Duration
}

// Default configuration
var defaultConfig = ProcessorConfig{
	MaxConcurrent: 5,
	DelayBetween:  500 * time.Millisecond,
}

// Global state
var (
	processing   bool
	totalAccounts int
	processedAccounts int
	successAccounts int
	failedAccounts int
	inProgressAccounts int
	mutex       sync.Mutex
	successFile string
	failFile    string
)

// ProcessAccounts processes a batch of accounts
func ProcessAccounts(accounts []*models.Account) {
	mutex.Lock()
	if processing {
		mutex.Unlock()
		logger.LogGeneralJSON(logger.LogLevelWarn, "Already processing accounts", nil)
		return
	}
	processing = true
	
	// Reset counters
	totalAccounts = len(accounts)
	processedAccounts = 0
	successAccounts = 0
	failedAccounts = 0
	inProgressAccounts = 0
	successFile = ""
	failFile = ""
	mutex.Unlock()

	// Log processing start
	logger.LogGeneralJSON(logger.LogLevelInfo, fmt.Sprintf("Starting to process %d accounts", totalAccounts), nil)
	logProgress()

	// Use a semaphore to limit concurrency
	sem := make(chan bool, defaultConfig.MaxConcurrent)
	var wg sync.WaitGroup

	// Process accounts
	for _, account := range accounts {
		// Rate limiting
		sem <- true
		wg.Add(1)

		// Process each account in a goroutine
		go func(acc *models.Account) {
			defer func() {
				<-sem  // Release semaphore
				wg.Done()
			}()

			mutex.Lock()
			inProgressAccounts++
			mutex.Unlock()

			processAccount(acc)

			mutex.Lock()
			inProgressAccounts--
			processedAccounts++
			
			if acc.Status == models.StatusSuccess {
				successAccounts++
			} else if acc.Status == models.StatusFailed {
				failedAccounts++
			}
			mutex.Unlock()

			logProgress()

			// Add delay between accounts
			time.Sleep(defaultConfig.DelayBetween)
		}(account)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Save results to Excel files
	saveResults(accounts)

	mutex.Lock()
	processing = false
	mutex.Unlock()

	logger.LogGeneralJSON(logger.LogLevelInfo, "All accounts processed", map[string]interface{}{
		"total":    totalAccounts,
		"success":  successAccounts,
		"failed":   failedAccounts,
	})
}

// processAccount simulates processing a single account
// In a real implementation, this would actually log into the account
func processAccount(account *models.Account) {
	username := account.Username

	// Log start of processing
	logger.LogAccountProcessJSON(
		logger.LogLevelInfo,
		fmt.Sprintf("Bắt đầu xử lý tài khoản: %s", username),
		username,
		"start",
		0,
	)

	account.Status = models.StatusProcessing

	// Simulate login process
	// Step 1: Initialize
	logger.LogAccountStepJSON(
		logger.LogLevelInfo,
		"Bước 1: Khởi tạo phiên đăng nhập",
		username,
		"init_session",
		1,
	)
	time.Sleep(time.Duration(500+rand.Intn(500)) * time.Millisecond)

	// Step 2: Sending credentials
	logger.LogAccountStepJSON(
		logger.LogLevelInfo,
		"Bước 2: Gửi thông tin đăng nhập",
		username,
		"send_credentials",
		2,
	)
	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

	// Randomly determine if login is successful (for demo purposes)
	// In real implementation, this would check the actual login result
	loginSuccess := rand.Intn(100) < 70 // 70% success rate

	if loginSuccess {
		// Step 3: Get account info
		logger.LogAccountStepJSON(
			logger.LogLevelInfo,
			"Bước 3: Lấy thông tin tài khoản",
			username,
			"fetch_account_info",
			3,
		)
		time.Sleep(time.Duration(500+rand.Intn(800)) * time.Millisecond)

		// Step 4: Get transaction history
		logger.LogAccountStepJSON(
			logger.LogLevelInfo,
			"Bước 4: Lấy lịch sử giao dịch",
			username,
			"fetch_transactions",
			4,
		)
		time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

		// Generate random data for demonstration
		balance := float64(rand.Intn(10000000)) / 100
		lastDeposit := float64(rand.Intn(1000000)) / 100
		depositTime := time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour).Format("2006-01-02 15:04:05")
		depositTxCode := fmt.Sprintf("D%d", rand.Int63n(10000000000000000000))

		// Simulate transaction
		logger.LogTransactionJSON(
			logger.LogLevelInfo,
			"Giao dịch gần nhất",
			username,
			depositTxCode,
			depositTime,
			1, // Deposit type
			lastDeposit,
			lastDeposit, // Balance after deposit
			true,
		)

		// Update account data
		account.Balance = balance
		account.LastDeposit = lastDeposit
		account.DepositTime = depositTime
		account.DepositTxCode = depositTxCode
		account.Status = models.StatusSuccess

		// Log success
		logger.LogAccountResultJSON(
			logger.LogLevelInfo,
			"✓ TÀI KHOẢN ĐĂNG NHẬP THÀNH CÔNG",
			username,
			true,
			balance,
			lastDeposit,
			depositTime,
			depositTxCode,
		)
	} else {
		// Login failed
		errorCode := "AUTH_FAILED"
		errorDetails := "Thông tin đăng nhập không chính xác"
		
		if rand.Intn(100) < 30 {
			errorCode = "CAPTCHA_REQUIRED"
			errorDetails = "Yêu cầu xác thực CAPTCHA"
		}

		account.Status = models.StatusFailed
		account.ErrorCode = errorCode
		account.ErrorDetails = errorDetails

		logger.LogErrorJSON(
			logger.LogLevelError,
			"✗ ĐĂNG NHẬP THẤT BẠI",
			username,
			errorCode,
			errorDetails,
		)
	}
}

// logProgress logs the current progress
func logProgress() {
	mutex.Lock()
	defer mutex.Unlock()

	logger.LogProgressJSON(
		logger.LogLevelInfo,
		"Tiến trình xử lý",
		processedAccounts,
		totalAccounts,
		inProgressAccounts,
		successAccounts,
		failedAccounts,
	)
}

// saveResults saves the processed accounts to Excel files
func saveResults(accounts []*models.Account) {
	// Save successful accounts
	successFilename, err := models.SaveAccountsToExcel(accounts, true)
	if err != nil {
		logger.LogGeneralJSON(logger.LogLevelError, fmt.Sprintf("Error saving success file: %s", err.Error()), nil)
	} else {
		mutex.Lock()
		successFile = successFilename
		mutex.Unlock()
	}

	// Save failed accounts
	failFilename, err := models.SaveAccountsToExcel(accounts, false)
	if err != nil {
		logger.LogGeneralJSON(logger.LogLevelError, fmt.Sprintf("Error saving fail file: %s", err.Error()), nil)
	} else {
		mutex.Lock()
		failFile = failFilename
		mutex.Unlock()
	}

	// Log result files
	logger.LogResultFilesJSON(
		logger.LogLevelInfo,
		"File kết quả đã được tạo",
		successFile,
		failFile,
	)
}