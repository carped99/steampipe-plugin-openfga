# Steampipe ACL Plugin - Standalone PostgreSQL FDW

이 가이드는 Steampipe ACL 플러그인을 standalone PostgreSQL Foreign Data Wrapper로 배포하고 사용하는 방법을 설명합니다.

## 개요

Standalone FDW 모드에서는:
- PostgreSQL 데이터베이스에 Steampipe FDW extension이 설치됩니다
- ACL 플러그인이 PostgreSQL에서 직접 쿼리 가능한 외부 테이블로 노출됩니다
- 별도의 Steampipe CLI 없이 표준 PostgreSQL 클라이언트로 접근 가능합니다

## 빌드 및 실행

### Docker Compose 사용 (권장)

```bash
# 1. OpenFGA store 생성
curl -X POST http://localhost:8080/stores \
  -H "Content-Type: application/json" \
  -d '{"name":"acl-demo"}' | jq -r '.id'
# 출력: 01K9Y2QSETQJE22F1BNEJ3ZWTM

# 2. Store ID를 환경변수로 설정
export OPENFGA_STORE_ID=01K9Y2QSETQJE22F1BNEJ3ZWTM

# 3. Docker Compose로 시작
docker-compose -f docker/docker-compose-standalone.yaml up -d

# 4. PostgreSQL에 연결
psql -h localhost -U steampipe -d steampipe
# Password: steampipe
```

### 수동 Docker 빌드

```bash
# 1. Standalone FDW 이미지 빌드
docker build -f Dockerfile-standalone -t steampipe-acl-fdw .

# 2. 컨테이너 실행
docker run -d \
  --name steampipe-postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=steampipe \
  -e OPENFGA_API_URL=http://openfga:8080 \
  -e OPENFGA_STORE_ID=01K9Y2QSETQJE22F1BNEJ3ZWTM \
  steampipe-acl-fdw

# 3. 연결
psql -h localhost -U steampipe -d steampipe
```

## 사용 방법

### PostgreSQL 연결 정보

```
Host:     localhost
Port:     5432
Database: steampipe
User:     steampipe
Password: steampipe
```

### SQL 쿼리 예시

#### 1. 기본 권한 체크

```sql
-- Alice가 document:test-doc-1을 볼 수 있는지 확인
SELECT allowed
FROM acl.sys_acl_permission
WHERE subject_type = 'user'
  AND subject_id = 'alice'
  AND relation = 'viewer'
  AND object_type = 'doc'
  AND object_id = 'test-doc-1';
```

#### 2. Helper 함수 사용

```sql
-- Helper 함수로 간단하게 체크
SELECT acl.check_permission(
  'user',       -- subject_type
  'alice',      -- subject_id
  'viewer',     -- relation
  'doc',        -- object_type
  'test-doc-1'  -- object_id
);
-- 결과: true 또는 false
```

#### 3. 애플리케이션 통합

```sql
-- 애플리케이션 테이블과 조인
SELECT
  d.id,
  d.title,
  acl.check_permission('user', 'alice', 'viewer', 'doc', d.id::text) as can_view
FROM documents d
WHERE acl.check_permission('user', 'alice', 'viewer', 'doc', d.id::text) = true;
```

#### 4. 대량 권한 체크

```sql
-- 여러 사용자의 권한 한번에 체크
SELECT
  subject_id,
  object_id,
  allowed
FROM acl.sys_acl_permission
WHERE subject_type = 'user'
  AND subject_id = ANY(ARRAY['alice', 'bob', 'charlie'])
  AND relation = 'viewer'
  AND object_type = 'doc'
  AND object_id = 'test-doc-1';
```

## 애플리케이션 통합

### Node.js (pg)

```javascript
const { Client } = require('pg');

const client = new Client({
  host: 'localhost',
  port: 5432,
  database: 'steampipe',
  user: 'steampipe',
  password: 'steampipe'
});

await client.connect();

// 권한 체크
const result = await client.query(
  'SELECT acl.check_permission($1, $2, $3, $4, $5)',
  ['user', 'alice', 'viewer', 'doc', 'test-doc-1']
);

const hasPermission = result.rows[0].check_permission;
console.log('Has permission:', hasPermission);

await client.end();
```

### Python (psycopg2)

```python
import psycopg2

conn = psycopg2.connect(
    host="localhost",
    port=5432,
    database="steampipe",
    user="steampipe",
    password="steampipe"
)

cursor = conn.cursor()

# 권한 체크
cursor.execute(
    "SELECT acl.check_permission(%s, %s, %s, %s, %s)",
    ('user', 'alice', 'viewer', 'doc', 'test-doc-1')
)

has_permission = cursor.fetchone()[0]
print(f'Has permission: {has_permission}')

cursor.close()
conn.close()
```

### Go (pgx)

```go
package main

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5"
)

func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx,
        "postgres://steampipe:steampipe@localhost:5432/steampipe")
    if err != nil {
        panic(err)
    }
    defer conn.Close(ctx)

    var hasPermission bool
    err = conn.QueryRow(ctx,
        "SELECT acl.check_permission($1, $2, $3, $4, $5)",
        "user", "alice", "viewer", "doc", "test-doc-1",
    ).Scan(&hasPermission)

    if err != nil {
        panic(err)
    }

    fmt.Printf("Has permission: %v\n", hasPermission)
}
```

## 성능 고려사항

### 연결 풀링

프로덕션 환경에서는 연결 풀링을 사용하세요:

```bash
# PgBouncer 예시
docker run -d \
  --name pgbouncer \
  -p 6432:6432 \
  -e DB_HOST=steampipe-postgres \
  -e DB_PORT=5432 \
  -e DB_USER=steampipe \
  -e DB_PASSWORD=steampipe \
  edoburu/pgbouncer
```

### 쿼리 최적화

```sql
-- ❌ 비효율적: 필수 조건 누락
SELECT * FROM acl.sys_acl_permission;

-- ✅ 효율적: 모든 필수 조건 제공
SELECT * FROM acl.sys_acl_permission
WHERE subject_type = 'user'
  AND subject_id = 'alice'
  AND relation = 'viewer'
  AND object_type = 'doc'
  AND object_id = 'test-doc-1';
```

## 설정 변경

### OpenFGA 연결 정보 변경

```bash
# 1. 컨테이너 중지
docker-compose -f docker/docker-compose-standalone.yaml down

# 2. config/acl.spc 파일 수정
vim config/acl.spc

# 3. 재시작
docker-compose -f docker/docker-compose-standalone.yaml up -d
```

### PostgreSQL 설정 커스터마이징

`docker-compose-standalone.yaml` 파일에서 환경변수 수정:

```yaml
environment:
  - POSTGRES_DB=mydb           # 데이터베이스 이름
  - POSTGRES_USER=myuser       # 사용자 이름
  - POSTGRES_PASSWORD=mypass   # 비밀번호
```

## 트러블슈팅

### FDW extension이 설치되지 않음

```sql
-- extension 확인
SELECT * FROM pg_extension WHERE extname = 'steampipe_fdw';

-- 수동 설치
CREATE EXTENSION IF NOT EXISTS steampipe_fdw;
```

### 외부 테이블이 보이지 않음

```sql
-- 스키마 확인
\dn

-- 테이블 확인
\det acl.*

-- 스키마 다시 import
DROP SCHEMA IF EXISTS acl CASCADE;
CREATE SCHEMA acl;
IMPORT FOREIGN SCHEMA acl FROM SERVER steampipe_server INTO acl;
```

### OpenFGA 연결 오류

```bash
# 컨테이너 로그 확인
docker logs steampipe-postgres-fdw

# 설정 파일 확인
docker exec steampipe-postgres-fdw cat /home/steampipe/.steampipe/config/acl.spc

# OpenFGA 연결 테스트
docker exec steampipe-postgres-fdw curl http://openfga:8080/healthz
```

## 프로덕션 배포

### 1. 보안 설정

```yaml
# docker-compose-standalone.yaml
services:
  postgres-fdw:
    environment:
      - POSTGRES_PASSWORD=${DB_PASSWORD}  # 환경변수 사용
    volumes:
      - /path/to/secure/acl.spc:/home/steampipe/.steampipe/config/acl.spc:ro
```

### 2. 백업 설정

```bash
# PostgreSQL 백업
docker exec steampipe-postgres-fdw pg_dump -U steampipe steampipe > backup.sql

# 복원
docker exec -i steampipe-postgres-fdw psql -U steampipe steampipe < backup.sql
```

### 3. 모니터링

```sql
-- 활성 연결 확인
SELECT * FROM pg_stat_activity WHERE datname = 'steampipe';

-- FDW 통계
SELECT * FROM pg_stat_foreign_tables;
```

## 추가 리소스

- [Steampipe Documentation](https://steampipe.io/docs)
- [PostgreSQL FDW Documentation](https://www.postgresql.org/docs/current/postgres-fdw.html)
- [OpenFGA Documentation](https://openfga.dev/docs)
