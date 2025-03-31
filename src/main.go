package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// 전역 변수 - SAST 취약점: 하드코딩된 비밀번호
var (
	dbPath       = "./data.db"
	secretKey    = "super_secret_key_1234"
	adminUser    = "admin"
	adminPass    = "admin123"
	serverPort   = "8080"
	databaseConn *sql.DB
)

// User 구조체
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Item 구조체
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       float64 `json:"price"`
}

// Response 구조체
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// 데이터베이스 초기화
	initDB()
	defer databaseConn.Close()

	// API 라우트 설정
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/users", usersHandler)
	http.HandleFunc("/api/users/", userHandler)
	http.HandleFunc("/api/items", itemsHandler)
	http.HandleFunc("/api/items/", itemHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/exec", execHandler)
	http.HandleFunc("/api/files", fileHandler)

	// 서버 시작
	log.Printf("서버가 http://localhost:%s 에서 실행 중입니다", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}

// 데이터베이스 초기화 함수
func initDB() {
	var err error
	// 데이터베이스 파일이 없으면 생성
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}

	// 데이터베이스 연결
	databaseConn, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// 테이블 생성
	createTables()
	// 초기 데이터 삽입
	insertInitialData()
}

// 테이블 생성 함수
func createTables() {
	// 사용자 테이블 생성
	userTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		email TEXT,
		role TEXT
	);`

	_, err := databaseConn.Exec(userTable)
	if err != nil {
		log.Fatal(err)
	}

	// 아이템 테이블 생성
	itemTable := `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price REAL
	);`

	_, err = databaseConn.Exec(itemTable)
	if err != nil {
		log.Fatal(err)
	}
}

// 초기 데이터 삽입 함수
func insertInitialData() {
	// 사용자 데이터가 없으면 삽입
	var count int
	err := databaseConn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		// SAST 취약점: 하드코딩된 비밀번호 사용
		_, err = databaseConn.Exec("INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)",
			adminUser, adminPass, "admin@example.com", "admin")
		if err != nil {
			log.Fatal(err)
		}

		_, err = databaseConn.Exec("INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)",
			"user1", "password123", "user1@example.com", "user")
		if err != nil {
			log.Fatal(err)
		}
	}

	// 아이템 데이터가 없으면 삽입
	err = databaseConn.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		_, err = databaseConn.Exec("INSERT INTO items (name, description, price) VALUES (?, ?, ?)",
			"Item 1", "This is item 1", 10.99)
		if err != nil {
			log.Fatal(err)
		}

		_, err = databaseConn.Exec("INSERT INTO items (name, description, price) VALUES (?, ?, ?)",
			"Item 2", "This is item 2", 20.99)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// 홈 핸들러
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	response := Response{
		Status:  "success",
		Message: "Welcome to Vulnerable Go API Server",
		Data:    map[string]string{"version": "1.0.0"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 사용자 목록 핸들러
func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 모든 사용자 조회
		rows, err := databaseConn.Query("SELECT id, username, password, email, role FROM users")
		if err != nil {
			sendErrorResponse(w, "데이터베이스 조회 오류", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role); err != nil {
				sendErrorResponse(w, "데이터 스캔 오류", http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}

		sendSuccessResponse(w, "사용자 목록 조회 성공", users)

	case http.MethodPost:
		// 새 사용자 생성
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			sendErrorResponse(w, "잘못된 요청 형식", http.StatusBadRequest)
			return
		}

		// SAST 취약점: 비밀번호 해싱 없음
		result, err := databaseConn.Exec("INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)",
			user.Username, user.Password, user.Email, user.Role)
		if err != nil {
			sendErrorResponse(w, "사용자 생성 실패", http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		user.ID = int(id)

		sendSuccessResponse(w, "사용자 생성 성공", user)

	default:
		sendErrorResponse(w, "지원하지 않는 메서드", http.StatusMethodNotAllowed)
	}
}

// 특정 사용자 핸들러
func userHandler(w http.ResponseWriter, r *http.Request) {
	// URL에서 사용자 ID 추출
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendErrorResponse(w, "잘못된 사용자 ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 특정 사용자 조회
		// SAST 취약점: SQL 인젝션 취약점
		query := fmt.Sprintf("SELECT id, username, password, email, role FROM users WHERE id = %d", id)
		row := databaseConn.QueryRow(query)

		var user User
		if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role); err != nil {
			if err == sql.ErrNoRows {
				sendErrorResponse(w, "사용자를 찾을 수 없음", http.StatusNotFound)
			} else {
				sendErrorResponse(w, "데이터베이스 조회 오류", http.StatusInternalServerError)
			}
			return
		}

		sendSuccessResponse(w, "사용자 조회 성공", user)

	case http.MethodPut:
		// 사용자 정보 업데이트
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			sendErrorResponse(w, "잘못된 요청 형식", http.StatusBadRequest)
			return
		}

		_, err := databaseConn.Exec("UPDATE users SET username = ?, password = ?, email = ?, role = ? WHERE id = ?",
			user.Username, user.Password, user.Email, user.Role, id)
		if err != nil {
			sendErrorResponse(w, "사용자 업데이트 실패", http.StatusInternalServerError)
			return
		}

		user.ID = id
		sendSuccessResponse(w, "사용자 업데이트 성공", user)

	case http.MethodDelete:
		// 사용자 삭제
		_, err := databaseConn.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			sendErrorResponse(w, "사용자 삭제 실패", http.StatusInternalServerError)
			return
		}

		sendSuccessResponse(w, "사용자 삭제 성공", nil)

	default:
		sendErrorResponse(w, "지원하지 않는 메서드", http.StatusMethodNotAllowed)
	}
}

// 아이템 목록 핸들러
func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 모든 아이템 조회
		rows, err := databaseConn.Query("SELECT id, name, description, price FROM items")
		if err != nil {
			sendErrorResponse(w, "데이터베이스 조회 오류", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var items []Item
		for rows.Next() {
			var item Item
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price); err != nil {
				sendErrorResponse(w, "데이터 스캔 오류", http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		sendSuccessResponse(w, "아이템 목록 조회 성공", items)

	case http.MethodPost:
		// 새 아이템 생성
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			sendErrorResponse(w, "잘못된 요청 형식", http.StatusBadRequest)
			return
		}

		result, err := databaseConn.Exec("INSERT INTO items (name, description, price) VALUES (?, ?, ?)",
			item.Name, item.Description, item.Price)
		if err != nil {
			sendErrorResponse(w, "아이템 생성 실패", http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		item.ID = int(id)

		sendSuccessResponse(w, "아이템 생성 성공", item)

	default:
		sendErrorResponse(w, "지원하지 않는 메서드", http.StatusMethodNotAllowed)
	}
}

// 특정 아이템 핸들러
func itemHandler(w http.ResponseWriter, r *http.Request) {
	// URL에서 아이템 ID 추출
	idStr := strings.TrimPrefix(r.URL.Path, "/api/items/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendErrorResponse(w, "잘못된 아이템 ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 특정 아이템 조회
		row := databaseConn.QueryRow("SELECT id, name, description, price FROM items WHERE id = ?", id)

		var item Item
		if err := row.Scan(&item.ID, &item.Name, &item.Description, &item.Price); err != nil {
			if err == sql.ErrNoRows {
				sendErrorResponse(w, "아이템을 찾을 수 없음", http.StatusNotFound)
			} else {
				sendErrorResponse(w, "데이터베이스 조회 오류", http.StatusInternalServerError)
			}
			return
		}

		sendSuccessResponse(w, "아이템 조회 성공", item)

	case http.MethodPut:
		// 아이템 정보 업데이트
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			sendErrorResponse(w, "잘못된 요청 형식", http.StatusBadRequest)
			return
		}

		_, err := databaseConn.Exec("UPDATE items SET name = ?, description = ?, price = ? WHERE id = ?",
			item.Name, item.Description, item.Price, id)
		if err != nil {
			sendErrorResponse(w, "아이템 업데이트 실패", http.StatusInternalServerError)
			return
		}

		item.ID = id
		sendSuccessResponse(w, "아이템 업데이트 성공", item)

	case http.MethodDelete:
		// 아이템 삭제
		_, err := databaseConn.Exec("DELETE FROM items WHERE id = ?", id)
		if err != nil {
			sendErrorResponse(w, "아이템 삭제 실패", http.StatusInternalServerError)
			return
		}

		sendSuccessResponse(w, "아이템 삭제 성공", nil)

	default:
		sendErrorResponse(w, "지원하지 않는 메서드", http.StatusMethodNotAllowed)
	}
}

// 로그인 핸들러
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "지원하지 않는 메서드", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		sendErrorResponse(w, "잘못된 요청 형식", http.StatusBadRequest)
		return
	}

	// SAST 취약점: 취약한 인증 메커니즘
	// DAST 취약점: 인증 우회 가능성
	row := databaseConn.QueryRow("SELECT id, username, role FROM users WHERE username = ? AND password = ?",
		credentials.Username, credentials.Password)

	var user struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, "잘못된 사용자 이름 또는 비밀번호", http.StatusUnauthorized)
		} else {
			sendErrorResponse(w, "로그인 처리 중 오류 발생", http.StatusInternalServerError)
		}
		return
	}

	// DAST 취약점: 민감한 정보 노출
	// 인증 토큰 생성 (보안에 취약한 방식)
	token := fmt.Sprintf("%x", md5.Sum([]byte(user.Username+secretKey)))

	sendSuccessResponse(w, "로그인 성공", map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// 명령 실행 핸들러
func execHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "지원하지 않는 메서드", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Command string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendErrorResponse(w, "잘못된 요청 형식", http.StatusBadRequest)
		return
	}

	// SAST 취약점: 명령어 인젝션 취약점
	// DAST 취약점: 원격 코드 실행 가능성
	cmd := exec.Command("sh", "-c", request.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("명령 실행 실패: %v", err), http.StatusInternalServerError)
		return
	}

	sendSuccessResponse(w, "명령 실행 성공", map[string]string{
		"output": string(output),
	})
}

// 파일 핸들러
func fileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "지원하지 않는 메서드", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		sendErrorResponse(w, "파일 이름이 필요합니다", http.StatusBadRequest)
		return
	}

	// DAST 취약점: 경로 순회 취약점
	file, err := os.Open(filename)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("파일을 열 수 없음: %v", err), http.StatusNotFound)
		return
	}
	defer file.Close()

	// 파일 내용을 응답으로 전송
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	io.Copy(w, file)
}

// 성공 응답 전송 함수
func sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// 오류 응답 전송 함수
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := Response{
		Status:  "error",
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
