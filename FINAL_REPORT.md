# DevSecOps 취약한 Go API 서버 - 최종 보고서

## 프로젝트 개요

이 프로젝트는 DevSecOps 파이프라인 포트폴리오를 위한 의도적으로 취약점이 포함된 Go API 백엔드 서버입니다. 이 서버는 SAST, SCA, DAST 각각에 대해 2-3개의 알려진 취약점을 포함하고 있으며, Semgrep, Trivy, OWASP ZAP을 사용하여 이러한 취약점을 탐지할 수 있습니다.

## GitHub 저장소

- **저장소 URL**: [https://github.com/dotlike13/devsecops-vulnerable-go-api](https://github.com/dotlike13/devsecops-vulnerable-go-api)

## 구현된 취약점 요약

### SAST 취약점 (Semgrep으로 탐지 가능)

1. **하드코딩된 비밀번호**
   - 소스 코드에 직접 비밀번호와 비밀 키가 하드코딩되어 있습니다.

2. **SQL 인젝션 취약점**
   - 사용자 입력을 직접 SQL 쿼리에 삽입하여 SQL 인젝션 공격에 취약합니다.

3. **명령어 인젝션 취약점**
   - 사용자 입력을 검증 없이 시스템 명령어로 실행하여 명령어 인젝션 공격에 취약합니다.

### SCA 취약점 (Trivy로 탐지 가능)

1. **오래된 베이스 이미지**
   - 오래된 버전의 Go와 Alpine Linux를 사용하여 알려진 취약점에 노출됩니다.

2. **취약한 버전의 OpenSSL**
   - 취약점이 있는 OpenSSL 버전을 명시적으로 설치합니다.

3. **취약한 버전의 라이브러리**
   - 알려진 취약점이 있는 버전의 Go 라이브러리를 사용합니다.

### DAST 취약점 (OWASP ZAP으로 탐지 가능)

1. **인증 우회 가능성**
   - 취약한 인증 메커니즘을 사용하여 인증을 우회할 가능성이 있습니다.

2. **민감한 정보 노출**
   - 로그인 응답에 민감한 정보가 포함되어 있습니다.

3. **경로 순회 취약점**
   - 사용자 입력을 검증 없이 파일 경로로 사용하여 경로 순회 공격에 취약합니다.

4. **원격 코드 실행 가능성**
   - 사용자 입력을 검증 없이 시스템 명령어로 실행하여 원격 코드 실행 공격에 취약합니다.

## 프로젝트 구조

```
devsecops-vulnerable-go-api/
├── Dockerfile          # 취약점이 포함된 Docker 이미지 설정
├── README.md           # 프로젝트 소개
├── VULNERABILITIES.md  # 취약점 및 테스트 방법 문서
├── FINAL_REPORT.md     # 최종 보고서
├── src/
│   ├── go.mod          # Go 모듈 정의
│   └── main.go         # 취약점이 포함된 Go API 서버 코드
└── todo.md             # 프로젝트 진행 상황
```

## 테스트 도구 및 방법

### Semgrep을 사용한 SAST 테스트

```bash
# Semgrep 설치
pip install semgrep

# Go 코드 스캔
semgrep --config=p/golang src/

# 특정 규칙으로 스캔
semgrep --config=p/golang-security src/
```

### Trivy를 사용한 SCA 테스트

```bash
# Trivy 설치
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin

# Docker 이미지 스캔
trivy image devsecops-vulnerable-go-api:latest

# 파일시스템 스캔
trivy fs .
```

### OWASP ZAP을 사용한 DAST 테스트

```bash
# OWASP ZAP 설치 (Docker 사용)
docker pull owasp/zap2docker-stable

# 기본 스캔
docker run -t owasp/zap2docker-stable zap-baseline.py -t http://localhost:8080

# API 스캔
docker run -t owasp/zap2docker-stable zap-api-scan.py -t http://localhost:8080/api -f openapi

# 전체 스캔
docker run -t owasp/zap2docker-stable zap-full-scan.py -t http://localhost:8080
```

## 빌드 및 실행 방법

### 로컬에서 빌드 및 실행

```bash
# Go 설치 필요
cd src
go mod tidy
go build -o main
./main
```

### Docker를 사용한 빌드 및 실행

```bash
# Docker 설치 필요
docker build -t devsecops-vulnerable-go-api .
docker run -p 8080:8080 devsecops-vulnerable-go-api
```

## API 엔드포인트

- `GET /` - 홈 페이지
- `GET /api/users` - 모든 사용자 조회
- `POST /api/users` - 새 사용자 생성
- `GET /api/users/:id` - 특정 사용자 조회
- `PUT /api/users/:id` - 사용자 정보 업데이트
- `DELETE /api/users/:id` - 사용자 삭제
- `GET /api/items` - 모든 아이템 조회
- `POST /api/items` - 새 아이템 생성
- `GET /api/items/:id` - 특정 아이템 조회
- `PUT /api/items/:id` - 아이템 정보 업데이트
- `DELETE /api/items/:id` - 아이템 삭제
- `POST /api/login` - 로그인
- `POST /api/exec` - 명령어 실행
- `GET /api/files` - 파일 다운로드

## 결론

이 프로젝트는 DevSecOps 파이프라인 포트폴리오를 위한 취약한 Go API 백엔드 서버를 제공합니다. 의도적으로 포함된 취약점들은 Semgrep, Trivy, OWASP ZAP과 같은 보안 도구를 사용하여 탐지할 수 있으며, 이를 통해 DevSecOps 파이프라인의 효과를 시연할 수 있습니다.

더 자세한 취약점 정보와 테스트 방법은 [VULNERABILITIES.md](https://github.com/dotlike13/devsecops-vulnerable-go-api/blob/master/VULNERABILITIES.md) 파일을 참조하세요.
