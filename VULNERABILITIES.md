# 취약점 문서

이 문서는 DevSecOps 파이프라인 포트폴리오를 위해 의도적으로 구현된 취약점들을 설명합니다.

## SAST (정적 애플리케이션 보안 테스트) 취약점

Semgrep을 사용하여 다음 취약점들을 탐지할 수 있습니다:

1. **하드코딩된 비밀번호**
   - 위치: `main.go` 파일의 전역 변수 섹션
   - 코드: `secretKey = "super_secret_key_1234"`, `adminPass = "admin123"`
   - 설명: 소스 코드에 직접 비밀번호와 비밀 키가 하드코딩되어 있어 코드에 접근할 수 있는 모든 사람이 이 정보를 볼 수 있습니다.
   - 테스트: Semgrep 규칙 `hardcoded-credentials`를 사용하여 탐지

2. **SQL 인젝션 취약점**
   - 위치: `main.go` 파일의 `userHandler` 함수
   - 코드: `query := fmt.Sprintf("SELECT id, username, password, email, role FROM users WHERE id = %d", id)`
   - 설명: 사용자 입력을 직접 SQL 쿼리에 삽입하여 SQL 인젝션 공격에 취약합니다.
   - 테스트: Semgrep 규칙 `go.lang.security.audit.database.string-formatted-query.string-formatted-query`를 사용하여 탐지

3. **명령어 인젝션 취약점**
   - 위치: `main.go` 파일의 `execHandler` 함수
   - 코드: `cmd := exec.Command("sh", "-c", request.Command)`
   - 설명: 사용자 입력을 검증 없이 시스템 명령어로 실행하여 명령어 인젝션 공격에 취약합니다.
   - 테스트: Semgrep 규칙 `go.lang.security.audit.dangerous-exec-command.dangerous-exec-command`를 사용하여 탐지

## SCA (소프트웨어 구성 분석) 취약점

Trivy를 사용하여 다음 취약점들을 탐지할 수 있습니다:

1. **오래된 베이스 이미지**
   - 위치: `Dockerfile`의 첫 번째 줄
   - 코드: `FROM golang:1.16-alpine3.13`
   - 설명: 오래된 버전의 Go와 Alpine Linux를 사용하여 알려진 취약점에 노출됩니다.
   - 테스트: Trivy를 사용하여 이미지 스캔 시 탐지

2. **취약한 버전의 OpenSSL**
   - 위치: `Dockerfile`의 RUN 명령어
   - 코드: `openssl=1.1.1k-r0`
   - 설명: 취약점이 있는 OpenSSL 버전을 명시적으로 설치합니다.
   - 테스트: Trivy를 사용하여 이미지 스캔 시 탐지

3. **취약한 버전의 라이브러리**
   - 위치: `Dockerfile`의 RUN 명령어 및 `go.mod` 파일
   - 코드: 
     ```
     go get github.com/mattn/go-sqlite3@v1.14.6
     go get github.com/gorilla/websocket@v1.4.2
     go get github.com/dgrijalva/jwt-go@v3.2.0
     ```
   - 설명: 알려진 취약점이 있는 버전의 Go 라이브러리를 사용합니다.
   - 테스트: Trivy를 사용하여 이미지 스캔 시 탐지

## DAST (동적 애플리케이션 보안 테스트) 취약점

OWASP ZAP을 사용하여 다음 취약점들을 탐지할 수 있습니다:

1. **인증 우회 가능성**
   - 위치: `main.go` 파일의 `loginHandler` 함수
   - 설명: 취약한 인증 메커니즘을 사용하여 인증을 우회할 가능성이 있습니다.
   - 테스트: OWASP ZAP을 사용하여 로그인 엔드포인트(`/api/login`)에 대한 퍼징 테스트 수행

2. **민감한 정보 노출**
   - 위치: `main.go` 파일의 `loginHandler` 함수
   - 설명: 로그인 응답에 민감한 정보가 포함되어 있습니다.
   - 테스트: OWASP ZAP을 사용하여 응답 헤더와 본문에서 민감한 정보 탐지

3. **경로 순회 취약점**
   - 위치: `main.go` 파일의 `fileHandler` 함수
   - 코드: `file, err := os.Open(filename)`
   - 설명: 사용자 입력을 검증 없이 파일 경로로 사용하여 경로 순회 공격에 취약합니다.
   - 테스트: OWASP ZAP을 사용하여 `/api/files?filename=../../../etc/passwd`와 같은 경로 순회 공격 시도

4. **원격 코드 실행 가능성**
   - 위치: `main.go` 파일의 `execHandler` 함수
   - 설명: 사용자 입력을 검증 없이 시스템 명령어로 실행하여 원격 코드 실행 공격에 취약합니다.
   - 테스트: OWASP ZAP을 사용하여 `/api/exec` 엔드포인트에 악의적인 명령어 전송

## 테스트 접근 방식

### Semgrep을 사용한 SAST 테스트

1. Semgrep 설치:
   ```bash
   pip install semgrep
   ```

2. Go 코드 스캔:
   ```bash
   semgrep --config=p/golang src/
   ```

3. 특정 규칙으로 스캔:
   ```bash
   semgrep --config=p/golang-security src/
   ```

### Trivy를 사용한 SCA 테스트

1. Trivy 설치:
   ```bash
   curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
   ```

2. Docker 이미지 스캔:
   ```bash
   trivy image devsecops-vulnerable-go-api:latest
   ```

3. 파일시스템 스캔:
   ```bash
   trivy fs .
   ```

### OWASP ZAP을 사용한 DAST 테스트

1. OWASP ZAP 설치 (Docker 사용):
   ```bash
   docker pull owasp/zap2docker-stable
   ```

2. 기본 스캔:
   ```bash
   docker run -t owasp/zap2docker-stable zap-baseline.py -t http://localhost:8080
   ```

3. API 스캔:
   ```bash
   docker run -t owasp/zap2docker-stable zap-api-scan.py -t http://localhost:8080/api -f openapi
   ```

4. 전체 스캔:
   ```bash
   docker run -t owasp/zap2docker-stable zap-full-scan.py -t http://localhost:8080
   ```
