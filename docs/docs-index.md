# Multi PocketBase UI - 문서 권한 인덱스

이 문서는 `docs/` 하위 설계 문서의 역할, 우선순위, 충돌 해소 규칙을 정의한다.
구현/리뷰/PR에서 문서 간 판단이 갈리면 이 문서를 먼저 따른다.

---

## 1) V0 기준 아키텍처 (정식)

- 제품/배포 이름: `pocketbase-multiview`
- 실행 바이너리: `pbmulti`
- 기본 모드: CLI(REPL)
- 핵심 모드: `pbmulti -ui`
- 배포 방식: prebuilt binary + Homebrew formula
- 지원 타깃(V0): macOS, Linux 바이너리

`ui-spec`의 Desktop/IPC 설계는 제품 기능 요구사항의 배경으로 유지하되,
런타임/배포 판단은 V0에서는 `deployment-cli-brew-spec`을 단일 기준으로 사용한다.

---

## 2) 문서별 역할

| 문서 | 역할 | 단일 기준(SOT) 범위 |
|---|---|---|
| `docs/docs-index.md` | 문서 권한/우선순위 정의 | 문서 충돌 해소 규칙 |
| `docs/deployment-cli-brew-spec.md` | 실행/배포 설계 | CLI 인터페이스, `-ui` 계약, 빌드/릴리스/Homebrew |
| `docs/ui-spec.md` | 제품 기능 명세 | 인스턴스/탐색/테이블/동기화/프리셋 동작 규칙 |
| `docs/spec-contracts.md` | 타입/상태 계약 | 상태 모델, 전이 규칙, 저장/런타임 경계 |
| `docs/ui-design-spec.md` | UI 시각/상호작용 규칙 | 토큰, 레이아웃, 상태별 디자인, A11y |
| `docs/structure-first-project-design.md` | 코드 구조 설계 원칙 | Primary Flow, Boundaries, Atom 분해, 테스트 구조 |

---

## 3) 우선순위 규칙

1. 문서 역할 충돌: `docs-index` 우선
2. 런타임/배포 충돌: `deployment-cli-brew-spec` 우선
3. 기능 동작 충돌: `ui-spec` 우선
4. 타입/상태 불일치: `spec-contracts` 우선
5. 시각 규칙 충돌: `ui-design-spec` 우선
6. 코드 구조 방식 충돌: `structure-first-project-design` 우선

---

## 4) 자주 발생하는 충돌의 공식 해석

1. `ui-spec`의 Desktop/IPC vs `deployment`의 Go CLI:
- V0 구현은 `deployment` 기준(`pbmulti`, `-ui`)으로 진행
- `ui-spec` 기능 요구는 유지하되 런타임 방식은 V0 배포 설계를 따름

2. Inspector 존재 여부:
- 제품 기능(본 구현): Inspector 존재 가능
- 정적 프리뷰: Inspector 제거 허용
- 실제 구현 범위 판단은 해당 단계의 목표 문서(V0 배포/프리뷰)에 따름

3. 저장 경로(LocalStorage vs OS 파일):
- 브라우저 프리뷰 설명은 UI 시뮬레이션 문맥
- 실행 앱 저장 정책은 `deployment` + `spec-contracts` 기준

---

## 5) 문서 변경 규칙

아래 항목을 변경하면 반드시 문서를 함께 수정한다.

1. CLI 플래그/명령/종료코드 변경:
- `deployment-cli-brew-spec.md`
- `docs-index.md`

2. 상태 필드/타입 변경:
- `spec-contracts.md`
- `ui-spec.md`의 관련 섹션

3. UI 동작 변경:
- `ui-spec.md`
- 필요 시 `ui-design-spec.md`

4. 코드 구조/파일 책임 변경:
- `structure-first-project-design.md`

---

## 6) 구현 시작 체크리스트

1. 현재 작업의 기준 문서 1개를 먼저 선언했는가
2. 변경이 다른 문서 SOT를 침범하는지 확인했는가
3. 충돌 시 우선순위를 `docs-index`로 해석했는가
4. 종료코드/에러포맷/플래그 계약을 테스트 항목에 포함했는가
5. `pbmulti -ui` 경로를 릴리스 게이트에 포함했는가

