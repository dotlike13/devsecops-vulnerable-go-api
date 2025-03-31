# SCA 취약점: 오래된 베이스 이미지 사용
FROM golang:1.16-alpine3.13

# SCA 취약점: 루트 사용자로 실행
USER root

# 작업 디렉토리 설정
WORKDIR /app

# 필요한 패키지 설치
# SCA 취약점: 취약한 버전의 패키지 설치
RUN apk add --no-cache --update \
    curl \
    git \
    openssh \
    bash \
    sqlite \
    # 취약한 버전의 OpenSSL 설치
    openssl=1.1.1k-r0

# 소스 코드 복사
COPY ./src /app/

# 의존성 설치
# SCA 취약점: 취약한 버전의 라이브러리 사용
RUN go get github.com/mattn/go-sqlite3@v1.14.6 && \
    go get github.com/gorilla/websocket@v1.4.2 && \
    go get github.com/dgrijalva/jwt-go@v3.2.0

# 애플리케이션 빌드
RUN go build -o main .

# 데이터베이스 디렉토리 생성
RUN mkdir -p /data && \
    chmod 777 /data

# 환경 변수 설정
ENV DB_PATH=/data/data.db
ENV SERVER_PORT=8080

# 포트 노출
EXPOSE 8080

# 애플리케이션 실행
CMD ["./main"]
