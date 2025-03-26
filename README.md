# Hệ Thống Kiểm Tra Đăng Nhập Tài Khoản

Ứng dụng web để upload file Excel chứa danh sách tài khoản, kiểm tra trạng thái đăng nhập và hiển thị kết quả theo thời gian thực.

## Tính năng

- Upload file Excel chứa danh sách tài khoản
- Hiển thị tiến trình xử lý theo thời gian thực
- Thống kê: tài khoản đúng, tài khoản sai, tổng tài khoản
- Tính toán chi phí (200 VND/tài khoản)
- Hiển thị thời gian xử lý
- Download kết quả tài khoản đúng/sai

## Công nghệ sử dụng

- **Frontend**: HTML, CSS, JavaScript, Tailwind CSS, Alpine.js
- **Backend**: Go (Golang)
- **WebSocket**: Truyền dữ liệu realtime
- **Xử lý file**: Xử lý file Excel (.xlsx, .xls)

## Cách cài đặt

1. Clone repository:
   ```
   git clone https://github.com/khahanv2/account-login-checker.git
   ```

2. Cài đặt dependencies:
   ```
   cd account-login-checker
   go mod download
   ```

3. Chạy ứng dụng:
   ```
   go run main.go
   ```

4. Truy cập ứng dụng tại: http://localhost:8080

## Cấu trúc dự án

```
account-login-checker/
├── cmd/
│   └── server/
│       └── main.go       # Entry point của ứng dụng
├── internal/
│   ├── config/           # Cấu hình ứng dụng
│   ├── handlers/         # HTTP Handlers
│   ├── logger/           # Hệ thống logging
│   ├── models/           # Data models
│   ├── storage/          # Xử lý lưu trữ
│   ├── utils/            # Utilities
│   └── websocket/        # WebSocket server
├── static/
│   ├── css/              # CSS files
│   ├── js/               # JavaScript files
│   └── images/           # Images
├── templates/            # HTML templates
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Cách sử dụng

1. Truy cập ứng dụng tại http://localhost:8080
2. Upload file Excel chứa danh sách tài khoản
3. Xem tiến trình xử lý theo thời gian thực
4. Download kết quả sau khi hoàn thành

## Format file Excel

File Excel cần có các cột sau:
- Username: Tên đăng nhập
- Password: Mật khẩu

## Giấy phép

MIT License