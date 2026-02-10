# multi-pocketbase-ui

PocketBase Admin UI의 불편함(컬럼 과다, 멀티 인스턴스 비교 불편)을 해소하기 위한 개인 로컬용 UI 프로젝트.

- V0 기준 런타임: `pbmulti` CLI + `pbmulti -ui` 로컬 UI 서버
- 관리자(Admin) 로그인은 PocketBase Admin API 규격을 따른다
- 인증 정보(`token`, `adminUser`)는 메모리 세션으로만 유지하며 영구 저장하지 않는다
- 설정/뷰/워크스페이스 프리셋만 로컬 저장한다
- 문서 권한/우선순위: `docs/docs-index.md`
- 기획/UX 설계안: `docs/ui-spec.md`
- 타입/상태 전이 계약: `docs/spec-contracts.md`
- UI 디자인 상세: `docs/ui-design-spec.md`
- CLI/Brew 배포 설계: `docs/deployment-cli-brew-spec.md`
- Structure First 설계: `docs/structure-first-project-design.md`
- CSS 디자인 토큰: `styles/tokens.css`
- 미리보기 인덱스: `preview/index.html`
- 2패널 화면: `preview/collections-2.html`
- 3패널 화면: `preview/collections-3.html`
- 4패널 화면: `preview/collections-4.html`
