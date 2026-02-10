# Multi PocketBase UI - CLI/Brew 배포 구현 설계서 (V0)

이 문서는 `pocketbase-multiview`를 DuckDB CLI 유사 UX로 배포하기 위한 **실행 설계서**다.
문서 간 충돌 해석은 `docs/docs-index.md`를 따른다.
목표는 다음 두 가지를 동시에 만족하는 것이다.

1. `brew install pocketbase-multiview`로 설치 가능
2. `pbmulti`는 CLI 중심, `pbmulti -ui`로 UI 실행

---

## 0) 제품 원칙 (고정)

### 0.1 사용자 우선순위

1. 1차 사용자: 사람이 `pbmulti -ui`로 직접 사용하는 UI 흐름
2. 2차 사용자: Codex/자동화가 CLI로 호출하는 흐름

### 0.2 `-ui` 필수 계약

- `-ui`는 선택 기능이 아니라 **항상 제공되는 핵심 인터페이스**다.
- 메이저 버전 변경 전까지 `-ui` 플래그는 제거/이름변경 금지다.
- 모든 릴리스에서 `pbmulti -ui` 동작 검증을 필수로 수행한다.

---

## 1) 목표와 범위

### 1.1 목표

- 단일 바이너리 `pbmulti` 제공
- 기본 실행 모드는 CLI(REPL)
- `-ui` 옵션으로 로컬 UI 서버 실행
- Homebrew 설치/업데이트 경로 제공

### 1.2 V0 포함 범위

- `pbmulti` 기본 명령 구조
- `pbmulti -ui` 서버 실행 + 브라우저 오픈
- `pbmulti version`, `pbmulti help`
- 정적 프리뷰 리소스 내장 서빙(`preview/`, `styles/`)
- macOS/Linux 바이너리 릴리스 제공
- Release 바이너리 생성 및 Homebrew Formula 배포

### 1.3 V0 제외 범위

- 완전한 PocketBase 쿼리 DSL
- 원격 telemetry/analytics
- 자동 업데이트 에이전트
- 서버형 멀티유저 기능
- Ubuntu 전용 `.deb` 패키지/apt 저장소 운영
- `-json` 구조화 출력(추후 V1)

### 1.4 모호성 제거를 위한 고정 결정

- 기본 실행 모드는 REPL로 유지한다.
- 1차 사용자 안내를 위해 REPL 시작 시 첫 줄에 `Tip: pbmulti -ui`를 출력한다.
- `-ui`는 다른 실행 모드 플래그(`-c`, script 파일)와 동시 사용 금지다.
- 배포는 V0에서 prebuilt binary 방식만 사용한다(Homebrew source build 미사용).

---

## 2) 사용자 경험(UX) 계약

### 2.1 설치

```bash
brew tap jiseop121/tap
brew install pocketbase-multiview
```

### 2.2 실행

```bash
pbmulti
# CLI REPL 시작

pbmulti -ui
# UI 서버 실행 + 기본 브라우저 오픈

pbmulti -ui -port 18080
# UI 서버 포트 지정

pbmulti -ui -no-browser
# 브라우저 오픈 없이 서버만 실행

pbmulti version
# 버전 출력

pbmulti -c "version"
# 원샷 명령 실행 후 종료

pbmulti script.pbmulti
# 스크립트 파일 실행 후 종료
```

### 2.3 종료

- CLI 모드: `exit`, `quit`, `Ctrl+D`
- UI 모드: `Ctrl+C`

---

## 3) 실행 모드 정의

### 3.1 CLI 모드 (default)

- 실행 조건: 인자 없음
- 동작:
  - ASCII 배너 출력(선택)
  - 첫 줄 안내: `Tip: pbmulti -ui`
  - REPL 루프 시작
  - 기본 명령(`help`, `version`, `open`, `exit`) 지원

### 3.2 UI 모드 (`-ui`)

- 실행 조건: `-ui=true`
- 동작:
  - 로컬 HTTP 서버 구동
  - `preview/index.html` 렌더
  - 정적 리소스(`/preview`, `/styles`) 제공
  - 기본 URL: `http://127.0.0.1:<port>`

### 3.3 충돌/예외

- 포트 사용 중:
  - 즉시 종료
  - 오류 메시지에 사용 포트 명시
- 브라우저 오픈 실패:
  - 서버는 유지
  - 수동 접속 URL 출력
- `-ui` 자체 실패:
  - 릴리스 블로커로 간주
  - 해당 빌드는 배포하지 않음
- 모드 플래그 충돌(`-ui` + `-c`, `-ui` + script 파일, `-c` + script 파일):
  - 실행하지 않고 즉시 종료
  - 종료 코드 `2`
  - 오류 메시지 예시: `ERR_INVALID_ARGS: conflicting mode flags`

---

## 4) 기술 아키텍처

### 4.1 언어/런타임

- Go 1.22+
- 단일 정적 바이너리 빌드

### 4.2 모듈 구조

```text
cmd/pbmulti/main.go
internal/app/run.go
internal/cli/repl.go
internal/cli/commands.go
internal/ui/server.go
internal/ui/open_browser.go
internal/buildinfo/version.go
internal/fs/embed.go
```

### 4.3 리소스 내장

- Go `embed` 사용
- 포함 경로:
  - `preview/**`
  - `styles/**`

### 4.4 경로/파일 정책

- 인증 데이터 디스크 저장 금지
- 설정/프리셋 저장 경로(향후 기능 대비):
  - macOS: `~/Library/Application Support/pbmulti/`
  - Linux: `~/.config/pbmulti/`
- 상태 경로 생성 실패 시:
  - fallback 없이 즉시 종료
  - 종료 코드 `1`
  - 오류 메시지에 실패 경로를 포함

---

## 5) CLI 인터페이스 상세

### 5.1 전역 플래그

- `-ui` (bool, default false)
- `-port` (int, default 17620)
- `-host` (string, default `127.0.0.1`)
- `-no-browser` (bool, default false)
- `-c` (string, default empty, one-shot command)
- `-state-dir` (string, default OS 표준 경로)

### 5.2 서브커맨드

- `version`
- `help`
- `ui` (선택: `pbmulti ui` 별칭)

### 5.3 REPL 커맨드(V0 최소)

- `help`
- `version`
- `ui`
- `exit`

### 5.4 DuckDB 유사 실행 패턴

- 인자 없음: 인터랙티브 REPL 시작
- `-c "<command>"`: 단일 명령 실행 후 종료
- `pbmulti <script-file>`: 스크립트 실행 후 종료

파싱 우선순위:
1. `-ui`가 있으면 UI 모드 우선
2. `-c`가 있으면 one-shot 모드
3. 위치 인자 1개가 있으면 script 모드
4. 그 외 REPL 모드

### 5.5 출력 형식

- 성공: 단문 + 필요한 값
- 실패: `ERR_CODE: message`
- 디버그 모드 제외 stack trace 금지
- 정상 결과는 `stdout`, 오류는 `stderr`로 분리

### 5.6 종료 코드 규약

- `0`: 성공
- `1`: 런타임 실패(파일 I/O, 포트 바인딩, 서버 실행 실패 등)
- `2`: 인자/모드 충돌/명령 문법 오류
- `3`: 외부 의존 실패(향후 PocketBase 네트워크/인증 실패용 예약)

### 5.7 Script 모드 문법(`*.pbmulti`)

- 인코딩: UTF-8
- 1줄 1명령
- 빈 줄 허용
- `#`로 시작하는 줄은 주석
- multiline 문법 미지원
- 첫 실패에서 즉시 중단(fail-fast)
- 실패 출력 형식: `ERR_SCRIPT_LINE_<N>: <message>`

### 5.8 Codex/자동화 호출 규칙

- UI 사용 자동화: `pbmulti -ui -no-browser`로 서버만 기동 후 URL 접근
- 원샷 실행 자동화: `pbmulti -c "<command>"`
- 비대화형 스크립트: `pbmulti <script-file>`
- 자동화 실행 시 표준 에러(`stderr`)에만 오류를 기록하고, 종료 코드를 명확히 반환

---

## 6) UI 서버 상세

### 6.1 라우팅

- `/` -> `preview/index.html`
- `/preview/*` -> embedded preview assets
- `/styles/*` -> embedded styles assets
- `/api/*` -> V0 미지원(항상 404 반환)

### 6.2 보안

- host 기본값 `127.0.0.1` 고정
- CORS 비활성(동일 출처만)
- directory listing 금지

### 6.3 헬스체크

- `/healthz` -> `200 ok`

---

## 7) 빌드/릴리스/배포

### 7.1 릴리스 산출물

- `pbmulti_<version>_darwin_arm64.tar.gz`
- `pbmulti_<version>_darwin_amd64.tar.gz`
- `pbmulti_<version>_linux_arm64.tar.gz`
- `pbmulti_<version>_linux_amd64.tar.gz`
- `checksums.txt`

### 7.2 태그 규칙

- `v0.1.0` 형태 SemVer

### 7.3 GitHub Actions 파이프라인

1. tag push 트리거
2. Go 빌드(4개 타깃: darwin/linux x arm64/amd64)
3. tar.gz 패키징
4. SHA256 생성
5. GitHub Release 업로드
6. Homebrew tap formula 업데이트 PR 생성

### 7.4 Homebrew Formula 계약

- Formula 이름: `pocketbase-multiview`
- 설치 바이너리명: `pbmulti`
- `test do`:
  - `system "#{bin}/pbmulti", "version"`

예시(초안):

```ruby
class PocketbaseMultiview < Formula
  desc "PocketBase multi-instance CLI and UI launcher"
  homepage "https://github.com/jiseop121/multi-pocketbase-ui"
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.1.0/pbmulti_0.1.0_darwin_arm64.tar.gz"
      sha256 "REPLACE_MACOS_ARM64_SHA256"
    else
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.1.0/pbmulti_0.1.0_darwin_amd64.tar.gz"
      sha256 "REPLACE_MACOS_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.1.0/pbmulti_0.1.0_linux_arm64.tar.gz"
      sha256 "REPLACE_LINUX_ARM64_SHA256"
    else
      url "https://github.com/jiseop121/multi-pocketbase-ui/releases/download/v0.1.0/pbmulti_0.1.0_linux_amd64.tar.gz"
      sha256 "REPLACE_LINUX_AMD64_SHA256"
    end
  end

  version "0.1.0"

  def install
    bin.install "pbmulti"
  end

  test do
    assert_match "pbmulti", shell_output("#{bin}/pbmulti version")
  end
end
```

### 7.5 Release Workflow 초안

```yaml
name: release
on:
  push:
    tags:
      - "v*"

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: darwin
            goarch: arm64
          - goos: darwin
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: linux
            goarch: amd64
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o pbmulti ./cmd/pbmulti
      - run: tar -czf pbmulti_${GITHUB_REF_NAME#v}_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz pbmulti
```

---

## 8) Homebrew Tap 운영 설계

### 8.1 저장소

- 앱: `jiseop121/multi-pocketbase-ui`
- tap repo: `jiseop121/homebrew-tap`
- tap 이름: `jiseop121/tap`

### 8.2 Formula 위치

- `Formula/pocketbase-multiview.rb`

### 8.3 업데이트 흐름

1. 앱 repo에서 새 태그 릴리스
2. checksum 확보
3. tap repo formula의 아키텍처별 `url`, `sha256`, `version` 갱신
4. 머지 후 `brew upgrade pocketbase-multiview` 가능

---

## 9) 테스트 계획

### 9.1 로컬 수동 테스트

1. `go run ./cmd/pbmulti` -> REPL 진입 확인
2. REPL 첫 줄에 `Tip: pbmulti -ui` 표시 확인
3. `go run ./cmd/pbmulti version` -> 버전 출력
4. `go run ./cmd/pbmulti -c "version"` -> one-shot 실행 확인
5. `go run ./cmd/pbmulti ./example/script.pbmulti` -> script 실행 확인
6. `go run ./cmd/pbmulti -ui -no-browser` -> 서버 시작/URL 출력
7. `/healthz` 응답 확인
8. 브라우저에서 `/` 접속 시 preview 인덱스 렌더 확인
9. `go run ./cmd/pbmulti -ui -c "version"` -> 종료 코드 `2` 확인

### 9.2 CI 테스트

- `go test ./...`
- `go vet ./...`
- 크로스 빌드 smoke test
- `pbmulti -ui -no-browser` 실행 smoke test (필수)
- Linux 타깃(`linux/arm64`, `linux/amd64`) 빌드 성공 검증
- 모드 충돌 종료 코드(`2`) 검증

### 9.3 Homebrew 설치 테스트

1. tap 추가
2. formula 설치
3. `pbmulti version` 실행
4. `pbmulti -ui -no-browser` 실행

---

## 10) 작업 분해(WBS)

### 10.1 Milestone A: 실행 파일 뼈대

- `cmd/pbmulti/main.go` 추가
- 플래그/인자 파서 구현
- version/help 구현
- REPL/one-shot/script 모드 분기 구현

완료 기준:
- `pbmulti`, `pbmulti -c`, `pbmulti <file>` 3가지 모드 동작

### 10.2 Milestone B: UI 서버

- embed FS 구성
- 정적 라우팅 + healthz
- 브라우저 오픈 유틸

완료 기준:
- `pbmulti -ui`에서 index 렌더

### 10.3 Milestone C: 릴리스 자동화

- release workflow 추가
- checksum 생성
- 아티팩트 업로드
- macOS/Linux 타깃 동시 빌드

완료 기준:
- 태그 푸시 시 macOS/Linux 릴리스 파일 자동 생성

### 10.4 Milestone D: Homebrew 배포

- tap 저장소 준비
- formula 최초 등록
- 설치 검증

완료 기준:
- `brew install pocketbase-multiview` 성공

### 10.5 Milestone E: 문서/운영 정리

- 설치/업데이트/문제해결 가이드 작성
- 릴리스 체크리스트 작성
- 운영 사고 대응 규칙 정리(롤백/핫픽스)

완료 기준:
- 신규 사용자가 문서만으로 설치/실행 가능

---

## 11) 리스크 및 대응

### 11.1 리스크

- macOS 브라우저 오픈 명령 차이
- Homebrew formula sha mismatch
- UI 정적 파일 경로 깨짐

### 11.2 대응

- `-no-browser` 제공으로 실행 자체 보장
- release job에서 checksum 아티팩트 고정
- 라우팅 smoke test를 CI에 추가

---

## 12) 완료 정의(DoD)

- `brew install pocketbase-multiview`로 설치 가능
- `pbmulti` 기본 실행 시 REPL 진입
- `pbmulti -ui` 실행 시 UI 접근 가능
- `-ui` 플래그가 문서/코드/테스트에 모두 존재
- macOS/Linux 바이너리 릴리스 아티팩트가 함께 생성됨
- 모드 충돌 시 종료 코드 `2`가 일관되게 반환됨
- script 모드 fail-fast + 라인 번호 에러가 확인됨
- 인증 비영속 원칙 위반 없음
- 문서/명령/실동작이 일치

---

## 13) 후속(V1+) 확장 포인트

- PocketBase 실제 질의 명령(`query`, `instances`, `login`)
- Workspace/TableView CLI 관리 명령
- `pbmulti ui` 정식 서브커맨드 전환
- Linux 패키지 배포(apt/rpm) 검토
- `-json` 구조화 출력

### 13.1 인증 입력 보안 원칙(사전 고정)

- `login` 기능 도입 시 비밀번호 플래그 직접 입력(`--password <value>`)은 금지한다.
- `--password-stdin` 방식만 허용한다.

---

## 14) 구현 파일 상세 명세

### 14.1 `cmd/pbmulti/main.go`

- 역할: entrypoint
- 책임:
  - 플래그 파싱
  - 모드 결정
  - `internal/app` 실행 위임

### 14.2 `internal/app/run.go`

- 역할: 런타임 오케스트레이터
- 책임:
  - mode 분기(REPL/UI/one-shot/script)
  - 종료 코드 관리

### 14.3 `internal/cli/repl.go`

- 역할: REPL loop
- 책임:
  - 입력 프롬프트/명령 dispatch
  - 기본 명령 처리(help/version/exit/ui)

### 14.4 `internal/cli/script.go`

- 역할: script mode
- 책임:
  - 파일 로드
  - 라인 단위 명령 실행
  - 에러 라인 번호 출력

### 14.5 `internal/ui/server.go`

- 역할: UI HTTP 서버
- 책임:
  - static asset 핸들러
  - `/healthz` 제공
  - graceful shutdown

### 14.6 `internal/fs/embed.go`

- 역할: 임베디드 리소스 제공
- 책임:
  - `preview`, `styles` 포함
  - path sanitization

---

## 15) 운영 체크리스트

릴리스 전:
1. `go test ./...` 통과
2. `pbmulti -c "version"` 수동 확인
3. `pbmulti -ui -no-browser` 수동 확인
4. 아티팩트 SHA 검증

릴리스 후:
1. clean 환경에서 `brew install pocketbase-multiview`
2. `pbmulti version` 실행
3. `pbmulti -ui` 렌더 확인
4. `brew upgrade pocketbase-multiview` 시나리오 확인
