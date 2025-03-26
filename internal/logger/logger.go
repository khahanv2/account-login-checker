package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// LogType defines the type of log
type LogType string

const (
	LogTypeGeneral       LogType = "general"
	LogTypeAccountProcess LogType = "account_process"
	LogTypeAccountStep    LogType = "account_step"
	LogTypeAccountResult  LogType = "account_result"
	LogTypeProgress       LogType = "progress"
	LogTypeTransaction    LogType = "transaction"
	LogTypeError          LogType = "error"
	LogTypeResultFiles    LogType = "result_files"
)

// LogLevel defines the level of log
type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelWarn  LogLevel = "WARN"
)

// LogData represents the JSON log structure
type LogData struct {
	Timestamp string      `json:"timestamp"`
	Level     LogLevel    `json:"level"`
	Source    string      `json:"source,omitempty"`
	Message   string      `json:"message"`
	Type      LogType     `json:"type"`
	Data      interface{} `json:"data"`
}

// WebSocketServer manages WebSocket connections
type WebSocketServer struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	register  chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex     sync.Mutex
	upgrader  websocket.Upgrader
	active    bool
}

// Global WebSocket server instance
var wsServer *WebSocketServer
var serverMutex sync.Mutex

// Initialize WebSocket server
func initWebSocketServer() *WebSocketServer {
	if wsServer != nil {
		return wsServer
	}

	wsServer = &WebSocketServer{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		active:     true,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all connections
			},
		},
	}

	// Start handling connections
	go wsServer.run()
	return wsServer
}

// StartWebSocketServer starts the WebSocket server
func StartWebSocketServer(port string) {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	ws := initWebSocketServer()
	if !ws.active {
		ws.active = true
		go ws.run()
	}

	// Create HTTP handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsServer.serveWs(w, r)
	})

	// Start HTTP server
	go func() {
		addr := ":" + port
		fmt.Printf("WebSocket server started on %s\n", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			fmt.Printf("WebSocket server error: %s\n", err)
		}
	}()
}

// StopWebSocketServer stops the WebSocket server
func StopWebSocketServer() {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	if wsServer != nil {
		wsServer.active = false
		close(wsServer.broadcast)
		for client := range wsServer.clients {
			client.Close()
		}
		wsServer = nil
	}
}

// GetActiveClientCount returns the number of active WebSocket clients
func GetActiveClientCount() int {
	if wsServer == nil {
		return 0
	}
	wsServer.mutex.Lock()
	defer wsServer.mutex.Unlock()
	return len(wsServer.clients)
}

// run handles WebSocket operations
func (server *WebSocketServer) run() {
	for server.active {
		select {
		case client := <-server.register:
			server.mutex.Lock()
			server.clients[client] = true
			server.mutex.Unlock()

		case client := <-server.unregister:
			server.mutex.Lock()
			if _, ok := server.clients[client]; ok {
				delete(server.clients, client)
				client.Close()
			}
			server.mutex.Unlock()

		case message, ok := <-server.broadcast:
			if !ok {
				// Channel closed
				return
			}

			server.mutex.Lock()
			for client := range server.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(server.clients, client)
				}
			}
			server.mutex.Unlock()
		}
	}
}

// serveWs handles WebSocket requests from clients
func (server *WebSocketServer) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Error upgrading connection: %v\n", err)
		return
	}

	server.register <- conn

	// Handle client disconnection
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				server.unregister <- conn
				break
			}
		}
	}()
}

// sendLogToWebSocket sends log data to WebSocket clients
func sendLogToWebSocket(logData LogData) {
	if wsServer == nil || !wsServer.active {
		return
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		fmt.Printf("Error marshaling log data: %v\n", err)
		return
	}

	wsServer.broadcast <- jsonData
}

// LogGeneralJSON logs a general message
func LogGeneralJSON(level LogLevel, message string, data interface{}) {
	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeGeneral,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}

// LogAccountProcessJSON logs account processing start
func LogAccountProcessJSON(level LogLevel, message string, username string, step string, stepNumber int) {
	data := map[string]interface{}{
		"username":    username,
		"step":        step,
		"step_number": stepNumber,
	}

	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeAccountProcess,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}

// LogAccountStepJSON logs account processing step
func LogAccountStepJSON(level LogLevel, message string, username string, step string, stepNumber int) {
	data := map[string]interface{}{
		"username":    username,
		"step":        step,
		"step_number": stepNumber,
	}

	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeAccountStep,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}

// LogAccountResultJSON logs account processing result
func LogAccountResultJSON(level LogLevel, message string, username string, success bool,
	balance float64, lastDeposit float64, depositTime string, depositTxCode string) {
	
	data := map[string]interface{}{
		"username":      username,
		"success":       success,
		"balance":       balance,
		"last_deposit":  lastDeposit,
		"deposit_time":  depositTime,
		"deposit_txcode": depositTxCode,
	}

	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeAccountResult,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}

// LogProgressJSON logs progress information
func LogProgressJSON(level LogLevel, message string, processed int, total int, inProgress int, successCount int, failCount int) {
	var successRate float64 = 0
	if processed > 0 {
		successRate = float64(successCount) / float64(processed) * 100
	}

	var percentComplete float64 = 0
	if total > 0 {
		percentComplete = float64(processed) / float64(total) * 100
	}

	data := map[string]interface{}{
		"processed":       processed,
		"total":           total,
		"in_progress":     inProgress,
		"success_rate":    successRate,
		"success_count":   successCount,
		"fail_count":      failCount,
		"percent_complete": percentComplete,
	}

	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeProgress,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}

// LogTransactionJSON logs transaction information
func LogTransactionJSON(level LogLevel, message string, username string, txNumber string, txTime string, txType int, amount float64, balanceAfter float64, isLatestDeposit bool) {
	data := map[string]interface{}{
		"username":          username,
		"transaction_number": txNumber,
		"transaction_time":   txTime,
		"transaction_type":   txType,
		"amount":            amount,
		"balance_after":     balanceAfter,
		"is_latest_deposit": isLatestDeposit,
	}

	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeTransaction,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}

// LogErrorJSON logs error information
func LogErrorJSON(level LogLevel, message string, username string, errorCode string, details string) {
	data := map[string]interface{}{
		"username":   username,
		"error_code": errorCode,
		"details":    details,
	}

	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeError,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}

// LogResultFilesJSON logs result files information
func LogResultFilesJSON(level LogLevel, message string, successFile string, failFile string) {
	data := map[string]interface{}{
		"success_file": successFile,
		"fail_file":    failFile,
	}

	logData := LogData{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
		Type:      LogTypeResultFiles,
		Data:      data,
	}
	sendLogToWebSocket(logData)
}