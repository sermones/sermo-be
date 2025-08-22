# Sermo Backend

Go Fiber를 사용한 HTTP 백엔드 API 서버입니다.

## 구조

```
sermo-be/
├── cmd/server/          # 메인 서버 진입점
├── internal/            # 내부 패키지
│   ├── config/         # 설정 관리
│   ├── handlers/       # HTTP 핸들러 (도메인별 분리)
│   │   └── auth/       # 인증 관련 핸들러
│   ├── models/         # 내부 비즈니스 모델
│   └── routes/         # 라우터 설정
├── pkg/                # 외부에서 사용할 수 있는 패키지
└── go.mod              # Go 모듈 정의
```

## 핸들러 구조

- **도메인별 분리**: `handlers/auth/signup_handler.go`, `handlers/auth/login_handler.go`
- **함수별 분리**: 각 파일에 하나의 핸들러 함수만 정의
- **비즈니스 로직**: 핸들러에서 직접 처리

## API 엔드포인트

### 인증 (Auth)
- `POST /auth/signup` - 회원가입
- `POST /auth/login` - 로그인

### 기본
- `GET /` - 환영 메시지
- `GET /health` - 헬스체크

## 실행 방법

```bash
# 서버 실행
go run cmd/server/main.go

# 또는 빌드 후 실행
go build -o bin/server cmd/server/main.go
./bin/server
```

## 환경 변수

- `PORT`: 서버 포트 (기본값: 3000)
- `HOST`: 서버 호스트 (기본값: localhost)

## 테스트

```bash
# 전체 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./internal/handlers/auth
```

