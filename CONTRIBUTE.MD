# Sermo Backend - 기여 가이드라인

## 📋 프로젝트 개요

Sermo Backend는 Go 언어와 Fiber 프레임워크를 사용하여 구축된 RESTful API 서버입니다. JWT 기반 인증, PostgreSQL 데이터베이스, 그리고 Swagger 문서화를 지원합니다.

## 🏗️ 아키텍처 개요

```
sermo-be/
├── cmd/server/          # 메인 애플리케이션 진입점
├── internal/            # 내부 패키지 (비공개)
│   ├── config/         # 설정 관리
│   ├── handlers/       # HTTP 요청 핸들러
│   ├── middleware/     # 미들웨어
│   ├── models/         # 데이터 모델
│   └── routes/         # 라우팅 설정
├── pkg/                # 공개 패키지
│   ├── database/       # 데이터베이스 연결 및 관리
│   ├── jwt/            # JWT 토큰 관리
│   ├── r2/             # Cloudflare R2 스토리지 (준비 중)
│   ├── logger/         # 로깅 (준비 중)
│   └── utils/          # 유틸리티 함수 (준비 중)
└── docs/               # Swagger API 문서
```

## 📦 패키지 구조 상세 분석

### 1. cmd/server/
- **main.go**: 애플리케이션의 진입점
  - 설정 로드
  - 데이터베이스 연결
  - 미들웨어 설정
  - 라우터 설정
  - Graceful shutdown 처리

### 2. internal/config/
- **config.go**: 환경변수 기반 설정 관리
  - 서버 설정 (호스트, 포트)
  - 데이터베이스 설정 (PostgreSQL)
  - R2 설정 (Cloudflare R2 스토리지)

### 3. internal/models/
- **user.go**: 사용자 데이터 모델
  - UUID 기반 ID
  - 사용자명, 이름, 비밀번호
  - 생성/수정 시간 자동 관리

### 4. internal/handlers/
#### auth/
- **signup_handler.go**: 회원가입 처리
- **login_handler.go**: 로그인 처리

#### user/
- **profile_handler.go**: 사용자 프로필 조회

### 5. internal/middleware/
- **auth_middleware.go**: JWT 토큰 검증
- **database_middleware.go**: 데이터베이스 연결 주입
- **config_middleware.go**: 설정 주입

### 6. internal/routes/
- **routes.go**: 메인 라우터 설정
- **auth_routes.go**: 인증 관련 라우트
- **user_routes.go**: 사용자 관련 라우트
- **health_routes.go**: 헬스체크 엔드포인트
- **swagger_routes.go**: Swagger UI 라우트

### 7. pkg/
- **database/**: PostgreSQL 연결 및 GORM 설정
- **jwt/**: JWT 토큰 생성 및 검증
- **r2/**: Cloudflare R2 스토리지 (준비 중)
- **logger/**: 로깅 시스템 (준비 중)
- **utils/**: 유틸리티 함수 (준비 중)

## 🔄 Request/Response DTO 구조

### 인증 관련 DTO

#### SignUpRequest
```go
type SignUpRequest struct {
    Nickname string `json:"nickname"` // 3-20자
    Name     string `json:"name"`     // 1-100자
    Password string `json:"password"` // 최소 6자
}
```

#### SignUpResponse
```go
type SignUpResponse struct {
    Message string `json:"message"`
    UserID  string `json:"user_id"`
}
```

#### LoginRequest
```go
type LoginRequest struct {
    ID       string `json:"id"`
    Password string `json:"password"`
}
```

#### LoginResponse
```go
type LoginResponse struct {
    Token string `json:"token"`
    User  struct {
        ID       string `json:"id"`
        Username string `json:"username"`
    } `json:"user"`
}
```

### 사용자 관련 DTO

#### ProfileResponse
```go
type ProfileResponse struct {
    UUID      string `json:"uuid"`
    ID        string `json:"id"`
    Nickname  string `json:"nickname"`
    CreatedAt string `json:"created_at"`
}
```

## 🔐 미들웨어 구조 및 동작 방식

### 1. 인증 미들웨어 (AuthMiddleware)
- **위치**: `internal/middleware/auth_middleware.go`
- **기능**: JWT 토큰 검증 및 사용자 ID 추출
- **동작 순서**:
  1. Authorization 헤더에서 Bearer 토큰 추출
  2. JWT 토큰 검증
  3. 사용자 ID를 context에 저장
  4. 다음 핸들러로 진행

### 2. 데이터베이스 미들웨어 (DatabaseMiddleware)
- **위치**: `internal/middleware/database_middleware.go`
- **기능**: 데이터베이스 연결을 context에 주입
- **사용법**: `middleware.GetDB(c)`로 핸들러에서 접근

### 3. 설정 미들웨어 (ConfigMiddleware)
- **위치**: `internal/middleware/config_middleware.go`
- **기능**: 애플리케이션 설정을 context에 주입
- **사용법**: `middleware.GetConfig(c)`로 핸들러에서 접근

## 🗄️ 데이터베이스 모델 및 스키마

### User 모델
```go
type User struct {
    UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    ID        string    `json:"id" gorm:"type:varchar(20);uniqueIndex;not null"`
    Nickname  string    `json:"nickname" gorm:"type:varchar(100);not null"`
    Password  string    `json:"-" gorm:"type:varchar(255);not null"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
```

### 데이터베이스 설정
- **드라이버**: PostgreSQL (GORM)
- **연결 풀**: 최대 100개 연결, 최대 10개 유휴 연결
- **타임존**: Asia/Seoul
- **자동 마이그레이션**: 활성화

## 🛣️ 라우팅 구조 및 API 엔드포인트

### 공개 엔드포인트
- `GET /health` - 서버 상태 확인
- `GET /swagger/*` - Swagger UI
- `POST /auth/signup` - 회원가입
- `POST /auth/login` - 로그인

### 인증 필요 엔드포인트
- `GET /user/profile` - 사용자 프로필 조회

### 라우터 그룹 구조
```go
// 인증 라우터 그룹
authGroup := app.Group("/auth")
authGroup.Post("/signup", auth.SignUp)
authGroup.Post("/login", auth.Login)

// 사용자 라우터 그룹 (인증 필요)
userGroup := app.Group("/user", middleware.AuthMiddleware())
userGroup.Get("/profile", user.GetProfile)
```

## 📚 의존성 및 외부 라이브러리

### 핵심 프레임워크
- **Fiber v2**: 고성능 HTTP 프레임워크
- **GORM v1**: ORM 라이브러리
- **PostgreSQL**: 데이터베이스

### 인증 및 보안
- **JWT v5**: JSON Web Token
- **UUID**: 고유 식별자 생성

### 문서화
- **Swagger**: API 문서 자동 생성

### 개발 도구
- **CORS**: 크로스 오리진 리소스 공유
- **Logger**: HTTP 요청 로깅
- **Recover**: 패닉 복구

## 🚀 개발 가이드라인

### 1. 코드 구조 규칙
- **패키지명**: 소문자, 단일 단어
- **파일명**: snake_case
- **함수명**: PascalCase (공개), camelCase (비공개)
- **변수명**: camelCase

### 2. 핸들러 작성 규칙
```go
// 1. DTO 정의 (Request/Response)
type ExampleRequest struct {
    Field string `json:"field"`
}

// 2. Swagger 주석 작성
// @Summary 요약
// @Description 상세 설명
// @Tags 태그
// @Accept json
// @Produce json
// @Param request body ExampleRequest true "요청 데이터"
// @Success 200 {object} ExampleResponse
// @Failure 400 {object} map[string]interface{}
// @Router /path [method]

// 3. 핸들러 함수 구현
func ExampleHandler(c *fiber.Ctx) error {
    // 요청 파싱
    var req ExampleRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // 비즈니스 로직
    
    // 응답 반환
    return c.JSON(response)
}
```

### 3. 미들웨어 작성 규칙
```go
func ExampleMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 전처리 로직
        
        // 다음 핸들러 호출
        if err := c.Next(); err != nil {
            return err
        }
        
        // 후처리 로직
        return nil
    }
}
```

### 4. 에러 처리 규칙
- HTTP 상태 코드 적절히 사용
- 일관된 에러 응답 형식
- 로깅 및 모니터링 고려

### 5. 데이터베이스 사용 규칙
- **마이그레이션 관리**: `database.RunMigrations()` 함수로 중앙 집중식 관리
- 트랜잭션 적절히 사용
- N+1 쿼리 문제 방지
- **모델 생성 시 반드시 New 함수 사용**: `models.NewUser()`, `models.NewImage()` 등

### 6. 마이그레이션 관리
- **자동 실행**: 서버 시작 시 `database.RunMigrations()` 자동 실행
- **수동 실행**: `make migrate` 명령어로 별도 실행 가능
- **모델 추가**: 새로운 모델은 `pkg/database/database.go`의 `RunMigrations()` 함수에 추가

## 🔧 개발 환경 설정

### 필수 요구사항
- Go 1.25.0 이상
- PostgreSQL 12 이상
- Make (선택사항)

### 환경변수 설정
```bash
# 서버 설정
PORT=3000
HOST=0.0.0.0

# 데이터베이스 설정
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=sermo
DB_SSLMODE=disable



# R2 설정 (준비 중)
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_ENDPOINT=your_endpoint
R2_BUCKET=your_bucket
```

### 실행 방법
```bash
# 개발 모드
go run cmd/server/main.go

# 빌드 후 실행
go build -o bin/server cmd/server/main.go
./bin/server
```

## 📝 기여 프로세스

1. **이슈 생성**: 새로운 기능이나 버그 수정 요청
2. **브랜치 생성**: `feature/기능명` 또는 `fix/버그명` 형식
3. **코드 작성**: 위의 가이드라인 준수
4. **테스트**: 기능 테스트 및 기존 기능 영향 확인
5. **Pull Request**: 상세한 설명과 함께 PR 생성
6. **코드 리뷰**: 팀원들의 리뷰 및 피드백 반영
7. **머지**: 승인 후 main 브랜치로 머지

## 🧪 테스트 가이드라인

### 단위 테스트
- 핸들러 함수별 테스트 작성
- 모델 함수별 테스트 작성
- 유틸리티 함수별 테스트 작성

### 통합 테스트
- API 엔드포인트별 테스트
- 데이터베이스 연동 테스트
- 미들웨어 동작 테스트

### 테스트 실행
```bash
# 전체 테스트
go test ./...

# 특정 패키지 테스트
go test ./internal/handlers/auth

# 테스트 커버리지
go test -cover ./...
```

## 📚 추가 리소스

- [Fiber 공식 문서](https://docs.gofiber.io/)
- [GORM 공식 문서](https://gorm.io/docs/)
- [JWT 공식 문서](https://jwt.io/)
- [Swagger 공식 문서](https://swagger.io/docs/)

---

**참고**: 이 문서는 프로젝트의 현재 상태를 반영하며, 프로젝트 발전에 따라 지속적으로 업데이트됩니다.
