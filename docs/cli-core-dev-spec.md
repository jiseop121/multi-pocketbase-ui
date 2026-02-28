# Multi PocketBase UI - Track 1: CLI 핵심 기능 개발 명세 (`-ui` 제외)

이 문서는 Track 1 개발의 단일 기준 문서다.
Track 1 기간에는 `-ui` 서버 기능을 구현하지 않는다.

## 1) 목표

- `pbmulti`를 CLI 중심 도구로 먼저 안정화한다.
- REPL / one-shot / script 실행 경로를 완성한다.
- 종료 코드/오류 형식을 고정해 자동화 호출 안정성을 확보한다.

## 2) 범위

### 2.1 포함
- 기본 실행: `pbmulti` -> REPL
- one-shot: `pbmulti -c "<command>"`
- script: `pbmulti <script-file>`
- 로컬 대상 관리: `db add/list/remove`
- 로컬 superuser 관리: `superuser add/list/remove`
- PocketBase 조회(GET): `api collections/collection/records/record`
- 출력 포맷: `table`(기본), `csv`, `markdown`
- 서브커맨드: `version`, `help`
- 표준 출력/표준 오류 분리
- 종료 코드 규약(`0/1/2/3`)

### 2.2 제외
- `-ui` 실제 실행(HTTP 서버, 브라우저 오픈, 정적 리소스 서빙)
- `/healthz` 및 UI 라우팅
- UI 관련 통합 테스트

## 3) 명령 계약

### 3.1 실행 모드
- 인자 없음: REPL
- `-c "..."`: one-shot
- 위치 인자 1개: script

### 3.2 충돌 규칙
- `-c` + script 동시 사용 금지
- 모드 충돌 시 즉시 종료 코드 `2`
- `stderr`에는 친절한 영어 오류 메시지를 출력한다.

### 3.3 `-ui` 처리(Track 1 한정)
- `-ui`는 예약 플래그로 인식만 한다.
- `-ui` 호출 시 실행하지 않고 종료 코드 `2` 반환.
- 사용자 오류 메시지는 `Error: UI mode is not available in Track 1.` 형식을 따른다.
- REPL 내부 `ui` 명령도 동일 정책을 따른다.

## 4) 출력/오류/종료 코드

출력 원칙:
- 성공 결과는 `stdout`
- 오류는 `stderr`
- 오류는 `Error: <message>` 형식으로 출력한다.
- 가능하면 다음 줄에 `Hint: <next action>`를 추가한다.

종료 코드:
- `0`: 성공
- `1`: 런타임 실패(파일 I/O 등)
- `2`: 인자/모드/미지원 기능(`-ui`) 오류
- `3`: 외부 의존 실패(예약)

## 5) Script 모드 계약

- 인코딩: UTF-8
- 1줄 1명령
- 빈 줄 허용
- `#` 주석 허용
- 첫 실패에서 중단(fail-fast)
- 에러 형식: `Error: Script failed at line <N>: <message>`

## 6) Track 1 테스트 게이트

릴리스 전 필수:
1. `go test ./...`
2. `pbmulti` REPL 진입 확인
3. `pbmulti -c "version"` 성공
4. `pbmulti <script-file>` fail-fast 동작 확인
5. `pbmulti -ui`가 `Error: UI mode is not available in Track 1.` + 종료 코드 `2` 반환
6. 모드 충돌(`-c` + script)에서 종료 코드 `2` 반환
7. `db/superuser` 관리 명령의 필수 인자 검증
8. `api records`에서 `--page --per-page --sort --filter` 반영 확인
9. `table` 출력이 ASCII 표 + `N rows` 형태인지 확인
10. `csv|markdown`에서 `--out` 누락 시 종료 코드 `2` 확인
11. 기본 오류 출력에 `ERR_` 문자열 미노출 확인
12. 오류는 `stderr`, 성공은 `stdout` 분리 확인
13. 인증 실패/네트워크 실패가 종료 코드 `3`으로 매핑되는지 확인

## 7) 완료 정의 (Track 1)

아래 조건을 모두 만족하면 Track 1 완료:
1. REPL / one-shot / script 3경로가 문서 계약대로 동작
2. 종료 코드/오류 포맷이 일관됨
3. `-ui` 요청이 미구현 상태로 안전하게 거절됨
4. 자동화 호출에서 표준 출력/오류 분리가 보장됨

## 8) Track 2 인계 항목

Track 2 시작 시 아래를 이관한다.
- `-ui` 예약 플래그를 실제 UI 서버 실행으로 전환
- REPL `ui` 명령을 실제 UI 모드 진입으로 전환
- UI 라우팅/헬스체크/브라우저 오픈 계약 적용

Track 2 상세 기준은 `docs/ui-mode-dev-spec.md`를 따른다.

## 9) CLI 인터페이스 계약 (PocketBase 조회 전용)

### 9.1 용어
- `db`: PocketBase 서버 접속 대상
- `dbAlias`: `db` 식별 별칭
- `superuser`: 특정 `db`에 접속할 PocketBase superuser 계정
- `superuserAlias`: `superuser` 식별 별칭

### 9.2 공통 계약
- REPL / one-shot / script에서 동일 명령 문법을 사용한다.
- PocketBase API 호출은 `GET`만 허용한다.
- 생성/수정/삭제 요청은 미지원이다.
- 성공 결과는 `stdout`, 오류는 `stderr`에 출력한다.
- 오류 출력은 아래 형식을 따른다.
  - 1행: `Error: <plain English message>`
  - 2행(선택): `Hint: <next action>`
- 자동화/스크립트는 오류 문구 파싱이 아니라 종료 코드(`0/1/2/3`)를 기준으로 처리해야 한다.

### 9.3 실행 문법
- one-shot: `pbmulti -c "<command>"`
- script: `pbmulti <script-file>`
- REPL: `pbmulti` 실행 후 명령 입력
- REPL 내부에서는 `pbmulti` 접두사 없이 동일 서브커맨드를 사용한다.
  - 예: one-shot `pbmulti -c "db list"` == REPL `db list`

### 9.4 db 관리 명령
- `pbmulti db add --alias <dbAlias> --url <baseUrl>`
- `pbmulti db list`
- `pbmulti db remove --alias <dbAlias>`

검증 규칙:
- `dbAlias`는 대소문자 무시 기준 고유값이다.
- `baseUrl`은 유효한 `http://` 또는 `https://` URL이어야 한다.

### 9.5 superuser 관리 명령
- `pbmulti superuser add --db <dbAlias> --alias <superuserAlias> --email <email> --password <password>`
- `pbmulti superuser list --db <dbAlias>`
- `pbmulti superuser remove --db <dbAlias> --alias <superuserAlias>`

검증 규칙:
- superuser는 반드시 하나의 `db`에 종속된다.
- 동일 `db` 내 `superuserAlias`는 고유값이다.

### 9.6 Collection 조회 명령
- `pbmulti api collections --db <dbAlias> --superuser <superuserAlias> [--format <table|csv|markdown>] [--out <path>]`
- `pbmulti api collection --db <dbAlias> --superuser <superuserAlias> --name <collectionName> [--format <table|csv|markdown>] [--out <path>]`

### 9.7 Record 조회 명령
- `pbmulti api records --db <dbAlias> --superuser <superuserAlias> --collection <collectionName> [--page <n>] [--per-page <n>] [--sort <expr>] [--filter <expr>] [--format <table|csv|markdown>] [--out <path>]`
- `pbmulti api record --db <dbAlias> --superuser <superuserAlias> --collection <collectionName> --id <recordId> [--format <table|csv|markdown>] [--out <path>]`

쿼리 규칙:
- `--page`, `--per-page`, `--sort`, `--filter`는 PocketBase 공식 규칙을 따른다.
- 잘못된 인자/쿼리 값은 종료 코드 `2`를 반환한다.

### 9.8 출력 포맷 계약
- 기본 포맷은 `table`이다.
- `--format table`
  - `stdout`에 ASCII 테이블 출력
  - `--out` 사용 불가
- `--format csv|markdown`
  - `--out <path>` 필수
  - 지정 파일에 결과 저장
  - `stdout`에는 저장 요약 1줄 출력
- 다건 조회 시 아래 메타 정보를 마지막 줄에 출력한다.
  - `page`, `perPage`, `totalItems`, `totalPages`
- 결과가 없으면 빈 테이블과 `0 rows`를 출력한다.

