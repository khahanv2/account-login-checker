package models

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// Account represents a user account to be checked
type Account struct {
	Username    string
	Password    string
	Balance     float64
	LastDeposit float64
	DepositTime string
	DepositTxCode string
	Status      AccountStatus
	ErrorCode   string
	ErrorDetails string
}

// AccountStatus represents the status of an account
type AccountStatus int

const (
	StatusNotProcessed AccountStatus = iota
	StatusProcessing
	StatusSuccess
	StatusFailed
)

// Transaction represents a transaction in the account history
type Transaction struct {
	Username      string
	Number        string
	Time          string
	Type          int // 1 = deposit, 2 = withdraw, etc.
	Amount        float64
	BalanceAfter  float64
	IsLatestDeposit bool
}

// LoadAccountsFromExcel loads accounts from an Excel file
func LoadAccountsFromExcel(filePath string) ([]*Account, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets in Excel file")
	}

	// Read rows from the first sheet
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	// Skip header row and process data
	accounts := make([]*Account, 0)
	for i, row := range rows {
		// Skip header row
		if i == 0 {
			continue
		}

		// Skip empty rows
		if len(row) < 2 || row[0] == "" || row[1] == "" {
			continue
		}

		// Create account (username and password are required)
		account := &Account{
			Username: row[0],
			Password: row[1],
			Status:   StatusNotProcessed,
		}

		accounts = append(accounts, account)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no valid accounts found in Excel file")
	}

	return accounts, nil
}

// SaveAccountsToExcel saves accounts to an Excel file
func SaveAccountsToExcel(accounts []*Account, successful bool) (string, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	
	// Set header row
	headerRow := []string{"Username", "Password", "Balance", "Last Deposit", "Deposit Time", "Deposit Transaction", "Status", "Error"}
	for i, header := range headerRow {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cell, header)
	}

	// Filter accounts based on success status
	filteredAccounts := make([]*Account, 0)
	for _, account := range accounts {
		if successful && account.Status == StatusSuccess {
			filteredAccounts = append(filteredAccounts, account)
		} else if !successful && account.Status == StatusFailed {
			filteredAccounts = append(filteredAccounts, account)
		}
	}

	// Add data rows
	for i, account := range filteredAccounts {
		rowNum := i + 2 // Start from row 2 (after header)
		
		status := "Not Processed"
		switch account.Status {
		case StatusProcessing:
			status = "Processing"
		case StatusSuccess:
			status = "Success"
		case StatusFailed:
			status = "Failed"
		}

		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), account.Username)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), account.Password)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowNum), account.Balance)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", rowNum), account.LastDeposit)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", rowNum), account.DepositTime)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", rowNum), account.DepositTxCode)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", rowNum), status)
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", rowNum), account.ErrorCode)
	}

	// Auto-fit column width
	for i := range headerRow {
		colName := string('A' + i)
		width, _ := f.GetColWidth("Sheet1", colName)
		if width < 15 {
			f.SetColWidth("Sheet1", colName, colName, 15)
		}
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	var filePrefix string
	if successful {
		filePrefix = "success"
	} else {
		filePrefix = "fail"
	}
	filename := fmt.Sprintf("%s_%s.xlsx", filePrefix, timestamp)
	filepath := "results/" + filename

	// Save the file
	if err := f.SaveAs(filepath); err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return filename, nil
}