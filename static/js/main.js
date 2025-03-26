// Alpine.js Component - Account Checker
function accountChecker() {
    return {
        // UI State
        isDragging: false,
        selectedFile: null,
        processing: false,

        // Processing Status
        totalAccounts: 0,
        processedAccounts: 0,
        successAccounts: 0,
        failedAccounts: 0,
        percentComplete: 0,
        processingTime: 0,
        
        // Result files
        successFile: null,
        failFile: null,
        
        // Timer references
        timerInterval: null,
        startTime: null,
        
        // Account data
        accounts: [],
        accountStartTimes: {},
        
        // WebSocket
        socket: null,
        
        // Initialize component
        init() {
            this.connectWebSocket();
            this.setupTimers();
        },
        
        // Connect to WebSocket server
        connectWebSocket() {
            this.socket = new WebSocket(`ws://${window.location.host}/ws`);
            
            this.socket.onopen = () => {
                console.log('WebSocket connected');
                this.showAlert('Kết nối WebSocket thành công', 'success');
            };
            
            this.socket.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleWebSocketMessage(data);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };
            
            this.socket.onclose = () => {
                console.log('WebSocket connection closed');
                
                // Try to reconnect after 3 seconds
                setTimeout(() => {
                    this.connectWebSocket();
                }, 3000);
            };
            
            this.socket.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.showAlert('Lỗi kết nối WebSocket', 'error');
            };
        },
        
        // Setup timers
        setupTimers() {
            // Update processing time every second if processing
            setInterval(() => {
                if (this.processing && this.startTime) {
                    const now = new Date();
                    this.processingTime = Math.floor((now - this.startTime) / 1000);
                    
                    // Update account processing times
                    this.updateAccountProcessingTimes();
                }
            }, 1000);
        },
        
        // Handle WebSocket messages
        handleWebSocketMessage(log) {
            console.log('Received log:', log);
            
            switch(log.type) {
                case 'account_process':
                    this.startAccountProcess(log.data);
                    break;
                    
                case 'account_step':
                    this.updateAccountStep(log.data);
                    break;
                    
                case 'account_result':
                    this.updateAccountResult(log.data);
                    break;
                    
                case 'progress':
                    this.updateProgress(log.data);
                    break;
                    
                case 'transaction':
                    this.updateTransactionInfo(log.data);
                    break;
                    
                case 'error':
                    this.showAccountError(log.data);
                    break;
                    
                case 'result_files':
                    this.updateResultFiles(log.data);
                    break;
            }
        },
        
        // Start processing a new account
        startAccountProcess(data) {
            const username = data.username;
            
            // Check if account already exists
            const existingAccount = this.accounts.find(acc => acc.username === username);
            if (existingAccount) return;
            
            // Track start time for this account
            this.accountStartTimes[username] = new Date();
            
            // Add account to list
            this.accounts.push({
                username: username,
                status: 'processing',
                step: data.step || 'Bắt đầu',
                stepNumber: data.step_number || 0,
                balance: null,
                lastDeposit: null,
                depositTime: null,
                processingTime: '0s'
            });
            
            // Sort accounts by processing status
            this.sortAccounts();
        },
        
        // Update account processing step
        updateAccountStep(data) {
            const username = data.username;
            const account = this.accounts.find(acc => acc.username === username);
            
            if (account) {
                account.step = data.step || 'Đang xử lý';
                account.stepNumber = data.step_number || 0;
            }
        },
        
        // Update account result
        updateAccountResult(data) {
            const username = data.username;
            const account = this.accounts.find(acc => acc.username === username);
            
            if (account) {
                account.status = data.success ? 'success' : 'failed';
                
                if (data.success) {
                    account.balance = data.balance || 0;
                    account.lastDeposit = data.last_deposit || null;
                    account.depositTime = data.deposit_time || null;
                    account.depositTxCode = data.deposit_txcode || null;
                }
                
                // Calculate final processing time
                if (this.accountStartTimes[username]) {
                    const processingSeconds = Math.floor((new Date() - this.accountStartTimes[username]) / 1000);
                    account.processingTime = this.formatDuration(processingSeconds);
                    
                    // Remove from active timing
                    delete this.accountStartTimes[username];
                }
                
                // Sort accounts by status
                this.sortAccounts();
            }
        },
        
        // Show account error
        showAccountError(data) {
            const username = data.username;
            const account = this.accounts.find(acc => acc.username === username);
            
            if (account) {
                account.status = 'failed';
                account.errorCode = data.error_code || 'Unknown';
                account.errorDetails = data.details || '';
                
                // Calculate final processing time
                if (this.accountStartTimes[username]) {
                    const processingSeconds = Math.floor((new Date() - this.accountStartTimes[username]) / 1000);
                    account.processingTime = this.formatDuration(processingSeconds);
                    
                    // Remove from active timing
                    delete this.accountStartTimes[username];
                }
                
                // Sort accounts by status
                this.sortAccounts();
            }
        },
        
        // Update transaction information
        updateTransactionInfo(data) {
            const username = data.username;
            const account = this.accounts.find(acc => acc.username === username);
            
            if (account && data.is_latest_deposit) {
                account.lastDeposit = data.amount || 0;
                account.depositTime = data.transaction_time || '';
                account.depositTxCode = data.transaction_number || '';
            }
        },
        
        // Update progress information
        updateProgress(data) {
            this.totalAccounts = data.total || 0;
            this.processedAccounts = data.processed || 0;
            this.successAccounts = data.success_count || 0;
            this.failedAccounts = data.fail_count || 0;
            this.percentComplete = data.percent_complete || 0;
            
            // If processing completed
            if (this.percentComplete >= 100 && this.processing) {
                this.processing = false;
                this.showAlert('Xử lý tất cả tài khoản hoàn tất!', 'success');
            }
        },
        
        // Update result files
        updateResultFiles(data) {
            if (data.success_file) {
                this.successFile = data.success_file;
            }
            
            if (data.fail_file) {
                this.failFile = data.fail_file;
            }
        },
        
        // Update all account processing times
        updateAccountProcessingTimes() {
            const now = new Date();
            
            this.accounts.forEach(account => {
                if (account.status === 'processing' && this.accountStartTimes[account.username]) {
                    const processingSeconds = Math.floor((now - this.accountStartTimes[account.username]) / 1000);
                    account.processingTime = this.formatDuration(processingSeconds);
                }
            });
        },
        
        // Sort accounts by status (processing first, then failed, then success)
        sortAccounts() {
            this.accounts.sort((a, b) => {
                // Processing accounts first
                if (a.status === 'processing' && b.status !== 'processing') return -1;
                if (a.status !== 'processing' && b.status === 'processing') return 1;
                
                // Then by status (success, failed)
                if (a.status !== b.status) {
                    if (a.status === 'success') return -1;
                    if (b.status === 'success') return 1;
                }
                
                // Then by username
                return a.username.localeCompare(b.username);
            });
        },
        
        // Handle file selection from input
        handleFileSelect(event) {
            const file = event.target.files[0];
            if (file) {
                if (this.isExcelFile(file)) {
                    this.selectedFile = file;
                } else {
                    this.showAlert('Vui lòng chọn file Excel (.xlsx hoặc .xls)', 'error');
                    event.target.value = '';
                }
            }
        },
        
        // Handle file drop
        handleFileDrop(event) {
            this.isDragging = false;
            const file = event.dataTransfer.files[0];
            
            if (file) {
                if (this.isExcelFile(file)) {
                    this.selectedFile = file;
                    document.getElementById('file-input').files = event.dataTransfer.files;
                } else {
                    this.showAlert('Vui lòng chọn file Excel (.xlsx hoặc .xls)', 'error');
                }
            }
        },
        
        // Check if file is Excel
        isExcelFile(file) {
            return file.name.endsWith('.xlsx') || file.name.endsWith('.xls');
        },
        
        // Remove selected file
        removeFile() {
            this.selectedFile = null;
            document.getElementById('file-input').value = '';
        },
        
        // Start processing file
        startProcessing() {
            if (!this.selectedFile) {
                this.showAlert('Vui lòng chọn file Excel trước', 'warning');
                return;
            }
            
            if (this.processing) {
                return;
            }
            
            // Reset UI
            this.resetState();
            
            // Start processing
            this.processing = true;
            this.startTime = new Date();
            
            // Create form data
            const formData = new FormData();
            formData.append('file', this.selectedFile);
            
            // Upload file
            this.showAlert('Đang tải file lên máy chủ...', 'info');
            
            fetch('/upload', {
                method: 'POST',
                body: formData
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Lỗi khi tải file lên');
                }
                return response.json();
            })
            .then(data => {
                console.log('Upload successful:', data);
                this.showAlert('Đã tải file thành công. Bắt đầu xử lý tài khoản...', 'success');
            })
            .catch(error => {
                console.error('Error uploading file:', error);
                this.showAlert('Lỗi: ' + error.message, 'error');
                this.processing = false;
            });
        },
        
        // Download result file
        downloadFile(type) {
            const fileName = type === 'success' ? this.successFile : this.failFile;
            if (!fileName) return;
            
            window.location.href = `/download/${fileName}`;
        },
        
        // Reset state
        resetState() {
            this.accounts = [];
            this.accountStartTimes = {};
            this.totalAccounts = 0;
            this.processedAccounts = 0;
            this.successAccounts = 0;
            this.failedAccounts = 0;
            this.percentComplete = 0;
            this.processingTime = 0;
            this.successFile = null;
            this.failFile = null;
        },
        
        // Format currency (VND)
        formatCurrency(amount) {
            return new Intl.NumberFormat('vi-VN').format(amount) + 'đ';
        },
        
        // Format duration
        formatDuration(seconds) {
            if (seconds < 60) {
                return `${seconds}s`;
            } else if (seconds < 3600) {
                const minutes = Math.floor(seconds / 60);
                const remainingSeconds = seconds % 60;
                return `${minutes}m ${remainingSeconds}s`;
            } else {
                const hours = Math.floor(seconds / 3600);
                const minutes = Math.floor((seconds % 3600) / 60);
                const remainingSeconds = seconds % 60;
                return `${hours}h ${minutes}m ${remainingSeconds}s`;
            }
        },
        
        // Format time (for display)
        formatTime(seconds) {
            if (!seconds) return '00:00:00';
            
            const hours = Math.floor(seconds / 3600).toString().padStart(2, '0');
            const minutes = Math.floor((seconds % 3600) / 60).toString().padStart(2, '0');
            const secs = Math.floor(seconds % 60).toString().padStart(2, '0');
            
            return `${hours}:${minutes}:${secs}`;
        },
        
        // Show alert
        showAlert(message, type = 'info', autoClose = true) {
            const alertsContainer = document.getElementById('alerts-container');
            
            const alertEl = document.createElement('div');
            alertEl.className = `alert-${Date.now()} flex items-center p-4 mb-2 text-sm rounded-lg shadow-md`;
            
            // Set background color based on type
            switch (type) {
                case 'success':
                    alertEl.classList.add('bg-green-100', 'text-green-800');
                    break;
                case 'error':
                    alertEl.classList.add('bg-red-100', 'text-red-800');
                    break;
                case 'warning':
                    alertEl.classList.add('bg-yellow-100', 'text-yellow-800');
                    break;
                default:
                    alertEl.classList.add('bg-blue-100', 'text-blue-800');
            }
            
            // Add icon based on type
            let icon;
            switch (type) {
                case 'success':
                    icon = '<i class="fas fa-check-circle mr-2"></i>';
                    break;
                case 'error':
                    icon = '<i class="fas fa-times-circle mr-2"></i>';
                    break;
                case 'warning':
                    icon = '<i class="fas fa-exclamation-circle mr-2"></i>';
                    break;
                default:
                    icon = '<i class="fas fa-info-circle mr-2"></i>';
            }
            
            alertEl.innerHTML = `
                ${icon}
                <span>${message}</span>
                <button type="button" class="ml-auto -mx-1.5 -my-1.5 text-gray-500 rounded-lg p-1 hover:bg-gray-200 inline-flex items-center justify-center h-6 w-6" onclick="this.parentElement.remove()">
                    <i class="fas fa-times text-xs"></i>
                </button>
            `;
            
            alertsContainer.appendChild(alertEl);
            
            // Auto close after 5 seconds
            if (autoClose) {
                setTimeout(() => {
                    alertEl.classList.add('opacity-0', 'transition-opacity', 'duration-500');
                    setTimeout(() => {
                        if (alertEl.parentElement) {
                            alertEl.remove();
                        }
                    }, 500);
                }, 5000);
            }
            
            return alertEl;
        }
    };
}