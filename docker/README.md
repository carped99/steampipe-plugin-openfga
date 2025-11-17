# GMR PostgreSQL with OpenFGA FDW

PostGIS 기반 PostgreSQL 이미지에 OpenFGA Foreign Data Wrapper extension이 포함되어 있습니다.

## 빌드

```bash
# 1. 캐시를 사용한 빌드 (빠름, 개발용)
task build

# 2. make standalone부터 다시 빌드 (추천, 로컬 플러그인 변경 시)
task build:fresh

# 3. 완전히 새로 빌드 (모든 캐시 무시)
task build:clean

# 또는 직접 빌드
# 일반 빌드 (캐시 사용)
docker buildx build \
  --build-arg POSTGIS_TAG=17-3.5 \
  --platform linux/amd64 \
  -f docker/Dockerfile \
  -t gaia3d/gmr-postgis \
  .

# make standalone부터 캐시 무시 (CACHEBUST로 무효화 지점 지정)
docker buildx build \
  --build-arg POSTGIS_TAG=17-3.5 \
  --build-arg CACHEBUST=$(date +%s) \
  --platform linux/amd64 \
  -f docker/Dockerfile \
  -t gaia3d/gmr-postgis \
  .

# 모든 캐시 무시
docker buildx build --no-cache \
  --build-arg POSTGIS_TAG=17-3.5 \
  --platform linux/amd64 \
  -f docker/Dockerfile \
  -t gaia3d/gmr-postgis \
  .
```

**빌드 옵션:**
- `task build` - 모든 캐시 사용 (빠름, 로컬 개발용)
- `task build:fresh` - `make standalone`부터 다시 빌드 (추천, 로컬 플러그인 코드 변경 시)
- `task build:clean` - 모든 캐시 무시하고 처음부터 빌드 (완전 클린 빌드)

**캐시 무효화 범위:**
```
Dockerfile 레이어 순서:
1. git clone steampipe-postgres-fdw
2. go mod download
3. apt-get install (fdw-builder)
4. COPY Go binaries
5. COPY steampipe-postgres-fdw source
6. ← CACHEBUST 무효화 지점 (build:fresh)
7. RUN make standalone ← 여기부터 재실행
8. COPY extension files
9. COPY init script
```

- `build` - 모든 캐시 사용 ✓ (1~9 모두 캐시)
- `build:fresh` - CACHEBUST부터 무효화 ✗ (1~5 캐시, **6~9 재실행**)
- `build:clean` - 모든 캐시 무시 ✗ (1~9 모두 재실행)

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
3. (선택) 환경변수가 설정되어 있으면 기본 Foreign Server 자동 생성

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

**방법 1: 자동 생성 (환경변수 사용)**

Docker Compose에서 환경변수를 설정하면 자동으로 생성됩니다:

```yaml
environment:
  OPENFGA_ENDPOINT: openfga:8081                          # 필수
  OPENFGA_STORE_ID: 01KA0FSR3W39HES8PTRXA4PDYP            # 필수
  # OPENFGA_API_TOKEN: your-api-token                     # 선택사항
  # OPENFGA_AUTHORIZATION_MODEL_ID: 01HXXXXXXXXXXXXXXXXXX # 선택사항
```

**방법 2: 수동 생성 (권장 - 프로덕션 환경)**

```sql
-- OpenFGA 서버 연결 설정
CREATE SERVER openfga_server
  FOREIGN DATA WRAPPER openfga_fdw
  OPTIONS (
    endpoint 'openfga:8081',                              -- 필수
    store_id '01KA0FSR3W39HES8PTRXA4PDYP',                -- 필수
    api_token 'your-api-token',                           -- 선택사항
    authorization_model_id '01HXXXXXXXXXXXXXXXXXX'        -- 선택사항
  );

-- User Mapping 생성
CREATE USER MAPPING FOR postgres
  SERVER openfga_server;
```

**방법 3: 다중 서버 연결**

```sql
-- 개발 환경
CREATE SERVER openfga_dev
  FOREIGN DATA WRAPPER openfga_fdw
  OPTIONS (endpoint 'localhost:8081', store_id '01DEV...');

-- 프로덕션 환경
CREATE SERVER openfga_prod
  FOREIGN DATA WRAPPER openfga_fdw
  OPTIONS (endpoint 'prod.openfga.com:8081', store_id '01PROD...');
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

### PostgreSQL 설정

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `POSTGRES_USER` | `postgres` | PostgreSQL 사용자 |
| `POSTGRES_PASSWORD` | - | PostgreSQL 비밀번호 (필수) |
| `POSTGRES_DB` | `postgres` | 기본 데이터베이스 이름 |

### OpenFGA FDW 설정

| 변수 | 기본값 | 필수 | 설명 |
|------|--------|------|------|
| `STEAMPIPE_LOG_LEVEL` | `WARN` | ❌ | FDW 로그 레벨 (TRACE, DEBUG, INFO, WARN, ERROR) |
| `OPENFGA_ENDPOINT` | - | ✅* | OpenFGA 서버 주소 (자동 서버 생성용) |
| `OPENFGA_STORE_ID` | - | ✅* | OpenFGA Store ID (자동 서버 생성용) |
| `OPENFGA_API_TOKEN` | - | ❌ | OpenFGA API 토큰 (인증 필요 시) |
| `OPENFGA_AUTHORIZATION_MODEL_ID` | - | ❌ | Authorization Model ID (특정 모델 사용 시) |

\* OPENFGA_ENDPOINT와 OPENFGA_STORE_ID는 자동 서버 생성을 원할 때만 필수입니다.

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
