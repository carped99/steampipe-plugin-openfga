# Test Data Generator Scripts

## generate_testdata.go

대량의 테스트 데이터를 생성하는 스크립트입니다.

### 사용법

#### 대화형 입력 (Interactive Mode)

스크립트는 항상 대화형 모드로 실행되며, 순차적으로 입력을 요청합니다.

```bash
go run scripts/generate_testdata.go
```

**우선순위 (Priority):**
1. **사용자 입력값** (직접 입력한 값)
2. **환경변수** (입력 없이 엔터만 친 경우)
3. **기본값** (endpoint만, `localhost:8081`)
4. **오류** (Store ID가 최종적으로 없는 경우)

**실행 예시 1: 직접 입력**
```
Enter OpenFGA endpoint [localhost:8081]: api.example.com:8080
Enter OpenFGA Store ID: 01KA0FSR3W39HES8PTRXA4PDYP

=== OpenFGA Test Data Generator ===
Connection:
  Endpoint: api.example.com:8080
  Store ID: 01KA0FSR3W39HES8PTRXA4PDYP
...
```

**실행 예시 2: 환경변수 사용**
```bash
export OPENFGA_API_URL="prod.openfga.local:8081"
export OPENFGA_STORE_ID="01KA0FSR3W39HES8PTRXA4PDYP"

go run scripts/generate_testdata.go
```

실행 화면:
```
Enter OpenFGA endpoint [prod.openfga.local:8081]:     ← 엔터만 치면 환경변수 사용
Enter OpenFGA Store ID [01KA0FSR3W39HES8PTRXA4PDYP]:  ← 엔터만 치면 환경변수 사용

=== OpenFGA Test Data Generator ===
Connection:
  Endpoint: prod.openfga.local:8081  ← 환경변수에서 가져옴
  Store ID: 01KA0FSR3W39HES8PTRXA4PDYP  ← 환경변수에서 가져옴
...
```

**기본값:**
- Users: 100,000
- Docs: 50,000
- Folders: 20,000

#### 커스텀 데이터 크기

```bash
go run scripts/generate_testdata.go -users 200000 -docs 100000 -folders 50000
```

프롬프트에서 endpoint와 storeId를 입력받습니다 (환경변수 fallback 지원).

#### OpenFGA에 튜플 생성 (향후 구현)

```bash
go run scripts/generate_testdata.go -create-tuples
```

프롬프트에서 endpoint와 storeId를 입력받고 튜플을 생성합니다.

### 옵션

- `-users <number>`: 생성할 사용자 수 (기본값: 100000)
- `-docs <number>`: 생성할 문서 수 (기본값: 50000)
- `-folders <number>`: 생성할 폴더 수 (기본값: 20000)
- `-create-tuples`: OpenFGA에 튜플 생성 (향후 구현)

### 입력 우선순위

모든 입력값은 다음 우선순위로 처리됩니다:

**Endpoint:**
1. 사용자 직접 입력
2. 환경변수 `OPENFGA_API_URL`
3. 기본값 `localhost:8081`

**Store ID:**
1. 사용자 직접 입력
2. 환경변수 `OPENFGA_STORE_ID`
3. 없으면 오류 (필수 값)

### 환경변수

- `OPENFGA_API_URL`: OpenFGA 서버 주소
- `OPENFGA_STORE_ID`: OpenFGA Store ID (필수)
- `OPENFGA_AUTHORIZATION_MODEL_ID`: (선택) Authorization Model ID

### 성능

- **생성 속도**: 초당 약 230,000개 엔티티
- **10만 사용자 + 5만 문서 + 2만 폴더**: 약 0.7초

### 예제 출력

```
=== OpenFGA Test Data Generator ===
Configuration:
  Users: 100000
  Docs: 50000
  Folders: 20000

Generated 20000 folders
Generated 50000 docs
Generated 100000 users
Test data generation completed in 736ms

=== Generation Report ===
Total entities: 170000
  - Users: 100000
  - Docs: 50000
  - Folders: 20000
Generation time: 736.6727ms
Entities per second: 230767.34

Sample Users (first 5):
  - user:a1b2c3d4-e5f6-7890-1234-567890abcdef
  - user:b2c3d4e5-f6g7-8901-2345-678901bcdefg
  ...
```

## 테스트 실행

### 기본 테스트

```bash
go test ./openfga/... -v -run TestGenerateTestData
```

### 대용량 데이터 테스트

```bash
go test ./openfga/... -v -run TestGenerateLargeTestData -timeout 5m
```

### 벤치마크

```bash
go test ./openfga/... -bench=BenchmarkTestDataGeneration -benchtime=10s
```

## 주의사항

1. **메모리**: 10만개 이상 데이터 생성 시 약 50-100MB 메모리 사용
2. **OpenFGA 제한**: OpenFGA 서버의 Write API rate limit 확인 필요
3. **네트워크**: 대량 튜플 생성 시 네트워크 대역폭 고려
