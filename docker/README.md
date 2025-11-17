# GMR PostgreSQL with OpenFGA FDW

PostGIS 기반 PostgreSQL 이미지에 OpenFGA Foreign Data Wrapper extension이 포함되어 있습니다.

## 빌드

```bash
# 이미지 빌드
task build

# 또는 직접 빌드
docker buildx build \
  --build-arg POSTGIS_TAG=17-3.5 \
  --platform linux/amd64 \
  -f docker/Dockerfile \
  -t gaia3d/gmr-postgis \
  .
```

## 실행

```bash
# Docker Compose로 실행 (PostgreSQL + OpenFGA)
docker compose -f docker/docker-compose-gmr.yaml up -d

# PostgreSQL만 실행
docker run -d \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  --name gmr-postgres \
  gaia3d/gmr-postgis:latest
```

## Extension 자동 설치

컨테이너 시작 시 자동으로 다음이 실행됩니다:

1. `template_openfga` 템플릿 데이터베이스 생성
2. `template_openfga`와 기본 데이터베이스에 `openfga_fdw` extension 설치

## 사용 방법

### 1. 데이터베이스 접속

```bash
psql -h localhost -U postgres -d gmr
```

### 2. Extension 확인

```sql
-- Extension 설치 확인
\dx openfga_fdw

-- Extension 버전 확인
SELECT * FROM pg_available_extensions
WHERE name = 'openfga_fdw';
```

### 3. Foreign Server 생성

```sql
-- OpenFGA 서버 연결 설정
CREATE SERVER openfga_server
  FOREIGN DATA WRAPPER openfga_fdw
  OPTIONS (
    endpoint 'openfga:8081',
    store_id '01KA0FSR3W39HES8PTRXA4PDYP'
  );

-- User Mapping 생성
CREATE USER MAPPING FOR postgres
  SERVER openfga_server;
```

### 4. Foreign Table Import

```sql
-- 특정 테이블 import
IMPORT FOREIGN SCHEMA openfga
  LIMIT TO (acl_permission)
  FROM SERVER openfga_server
  INTO public;

-- 전체 테이블 import
IMPORT FOREIGN SCHEMA openfga
  FROM SERVER openfga_server
  INTO public;
```

### 5. 쿼리 실행

```sql
-- ACL 권한 조회
SELECT * FROM acl_permission
WHERE object_type = 'doc'
  AND relation = 'viewer'
LIMIT 10;

-- 특정 사용자의 권한 조회
SELECT * FROM acl_permission
WHERE subject_id = 'user-123';
```

## 새 데이터베이스에서 사용

`template_openfga`를 템플릿으로 사용하면 extension이 미리 설치된 데이터베이스를 생성할 수 있습니다:

```sql
-- Extension이 미리 설치된 데이터베이스 생성
CREATE DATABASE mydb TEMPLATE template_openfga;
```

## 환경 변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `POSTGRES_USER` | `postgres` | PostgreSQL 사용자 |
| `POSTGRES_PASSWORD` | - | PostgreSQL 비밀번호 (필수) |
| `POSTGRES_DB` | `postgres` | 기본 데이터베이스 이름 |

## 포트

- PostgreSQL: `5432`
- OpenFGA gRPC: `8081`
- OpenFGA HTTP: `8080`
- OpenFGA Playground: `3000`

## 볼륨

- `/var/lib/postgresql/data` - PostgreSQL 데이터 디렉토리

## 문제 해결

### Extension이 설치되지 않는 경우

```bash
# 컨테이너 로그 확인
docker logs gmr-postgres-openfga

# 초기화 스크립트 수동 실행
docker exec -it gmr-postgres-openfga bash
psql -U postgres -d gmr -c "CREATE EXTENSION openfga_fdw;"
```

### OpenFGA 연결 오류

```bash
# OpenFGA 서비스 상태 확인
docker logs openfga

# 네트워크 연결 확인
docker exec -it gmr-postgres-openfga ping openfga
```
