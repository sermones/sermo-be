# Sermo Backend API Documentation

## 📖 문서 보기

### 1. HTML 문서 (권장)
```bash
# 브라우저에서 docs/index.html 열기
open docs/index.html
```

### 2. Swagger UI
```bash
# 서버 실행 후 브라우저에서 접근
http://localhost:3000/swagger/
```

### 3. OpenAPI 스펙 파일
- `swagger.yaml` - YAML 형식
- `swagger.json` - JSON 형식

## 🔄 문서 업데이트

코드 변경 후 문서를 업데이트하려면:

```bash
# Swagger 문서 재생성
make swagger

# 또는 직접 실행
/Users/chungjung-mac-m4/.asdf/installs/golang/1.25.0/bin/swag init -g cmd/server/main.go -o docs
```

## 📝 Swagger 주석 작성법

### 핸들러 함수에 주석 추가
```go
// @Summary API 요약
// @Description API 상세 설명
// @Tags 태그명
// @Accept json
// @Produce json
// @Param request body RequestType true "요청 설명"
// @Success 200 {object} ResponseType
// @Failure 400 {object} ErrorResponse
// @Router /path [method]
func Handler(c *fiber.Ctx) error {
    // 핸들러 로직
}
```

### DTO 구조체에 주석 추가
```go
type RequestType struct {
    Field string `json:"field" example:"example_value" description:"필드 설명"`
}
```

## 🎨 Zudoku 스타일

현재 문서는 Redoc을 사용하여 Zudoku 스타일로 렌더링됩니다:
- 깔끔하고 현대적인 디자인
- 반응형 레이아웃
- 다크/라이트 테마 지원
- 검색 및 필터링 기능
