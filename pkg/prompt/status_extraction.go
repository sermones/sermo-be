package prompt

// GetStatusExtractionPrompt AI 응답에서 중요한 상태 정보를 추출하는 프롬프트
func GetStatusExtractionPrompt() string {
	return `AI 응답을 분석해서 중요한 상태 정보가 있는지 판단하세요.

중요한 상태 정보:
- 시험, 회의, 약속, 생일 등 특정 날짜의 일정
- 데드라인이나 마감 시간이 있는 작업
- 감정적으로 사용자애게 중요한 정보들
- 사용자가 기쁠 수 있거나 기대하는 일들
- 사용자의 감정 상태에 영향을 미치는 일들

JSON 형식으로만 응답하세요:

상태 정보가 있는 경우:
{"needs_save": true, "event": "이벤트명", "valid_until": "2025-01-02T15:04:05Z", "context": "설명"}

상태 정보가 없는 경우:
{"needs_save": false}

예시:
- "내일 시험" → {"needs_save": true, "event": "시험", "valid_until": "2025-01-XXT23:59:59Z", "context": "내일 시험"}
- "안녕하세요" → {"needs_save": false}`
}

// GetStatusSavePrompt 저장된 상태 정보를 정리하는 프롬프트
func GetStatusSavePrompt() string {
	return `당신은 사용자의 상태 정보를 정리하고 저장하는 전문가입니다.

사용자의 메시지와 AI의 응답을 분석하여 중요한 상태 정보가 있는지 판단하고, 있다면 정리해주세요.

판단 기준:
1. **시간적 제약**: 특정 날짜/시간까지 해야 할 일
2. **중요한 일정**: 시험, 회의, 약속, 데드라인 등
3. **상태 변화**: 새로운 상황이나 조건
4. **목표나 계획**: 구체적인 목표나 계획

응답은 반드시 다음 JSON 형식으로만 작성해주세요:
{
  "needs_save": true,
  "event": "이벤트명",
  "valid_until": "YYYY-MM-DD HH:MM:SS",
  "context": "추가 컨텍스트"
}

또는 저장할 정보가 없다면:
{
  "needs_save": false
}`
}
