# Steampipe ACL Plugin - Quick Start Guide

OpenFGA 권한을 PostgreSQL에서 직접 쿼리할 수 있는 Steampipe 플러그인입니다.

## 빠른 시작

### 옵션 1: Standalone PostgreSQL FDW (권장)

PostgreSQL에서 직접 OpenFGA 권한을 쿼리합니다.

```bash
# 1. Docker Compose로 실행
task standalone:up

# 2. PostgreSQL 연결
psql -h localhost -U steampipe -d steampipe
# Password: steampipe

# 3. 권한 쿼리
SELECT * FROM acl.sys_acl_permission
WHERE subject_type = 'user'
  AND subject_id = 'alice'
  AND relation = 'viewer'
  AND object_type = 'doc'
  AND object_id = 'test-doc-1';
```

**자세한 내용**: [STANDALONE_FDW.md](./STANDALONE_FDW.md)

### 옵션 2: Steampipe CLI

Steampipe CLI를 사용하여 쿼리합니다.

```bash
# 1. Steampipe 시작
task up

# 2. Steampipe 쿼리
docker exec -it steampipe-acl steampipe query \
  "SELECT * FROM acl.sys_acl_permission WHERE ..."
```

## 주요 기능

### 1. 실시간 권한 체크

```sql
-- 특정 사용자가 특정 문서를 볼 수 있는지 확인
SELECT allowed
FROM acl.sys_acl_permission
WHERE subject_type = 'user'
  AND subject_id = 'alice'
  AND relation = 'viewer'
  AND object_type = 'doc'
  AND object_id = 'my-document';
```

### 2. Helper 함수 (Standalone FDW only)

```sql
-- 간단한 함수 호출
SELECT acl.check_permission(
  'user', 'alice', 'viewer', 'doc', 'my-document'
);
```

### 3. 애플리케이션 통합

```javascript
// Node.js 예시
const result = await pg.query(
  'SELECT acl.check_permission($1, $2, $3, $4, $5)',
  ['user', 'alice', 'viewer', 'doc', 'my-document']
);
```

## 사용 가능한 명령어

```bash
# Standalone FDW
task standalone:up      # FDW 스택 시작
task standalone:down    # FDW 스택 중지
task standalone:logs    # 로그 확인

# Steampipe CLI
task build              # Docker 이미지 빌드
task run                # Steampipe CLI 실행
task up                 # Steampipe 서비스 시작

# 테스트
task test               # 단위 테스트
task test:integration   # 통합 테스트
```

## 설정

### OpenFGA 연결 정보

`config/acl.spc` 파일 수정:

```hcl
connection "acl" {
  plugin = "local/gaia3d/acl"

  api_url  = "http://openfga:8080"
  store_id = "YOUR_STORE_ID"
}
```

## 다음 단계

- [Standalone FDW 가이드](./STANDALONE_FDW.md) - PostgreSQL FDW 상세 가이드
- [테스트 가이드](./acl/README_TEST.md) - 테스트 실행 방법
- [OpenFGA 문서](https://openfga.dev/docs) - OpenFGA 설정 및 사용법

## 트러블슈팅

### OpenFGA 서버가 없는 경우

```bash
# OpenFGA 서버 시작
docker run -d -p 8080:8080 -p 8081:8081 \
  openfga/openfga run

# Store 생성
curl -X POST http://localhost:8080/stores \
  -H "Content-Type: application/json" \
  -d '{"name":"my-store"}'
```

### 연결 오류

```bash
# 로그 확인
task standalone:logs

# 컨테이너 상태 확인
docker ps
```

## 라이센스

See [LICENSE](./LICENSE)
