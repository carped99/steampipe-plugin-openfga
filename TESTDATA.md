# OpenFGA Test Data Generator

Faker를 사용한 대량 테스트 데이터 생성 도구

## 🚀 주요 기능

- **UUID 기반 ID**: 모든 엔티티는 UUID v4 형식의 고유 ID 사용
- **고성능**: 초당 **225,000개** 이상의 엔티티 생성
- **동시 처리**: Goroutine을 활용한 병렬 데이터 생성
- **대용량 지원**: 10만개 이상의 엔티티 생성 가능

## 📊 성능 벤치마크

| 데이터셋 | 엔티티 수 | 생성 시간 | 초당 엔티티 |
|---------|----------|----------|------------|
| Small | 1,700 | 10ms | 163,669 |
| Large | 170,000 | 753ms | 225,713 |

## 🛠️ 사용 방법

### 1. 기본 사용 (10만 + 5만 + 2만)

```bash
go run scripts/generate_testdata.go
```

출력:
```
=== OpenFGA Test Data Generator ===
Configuration:
  Users: 100000
  Docs: 50000
  Folders: 20000

Generated 170000 entities in 753ms

Sample Users (first 5):
  - user:d26cb748-2624-4af4-98d7-2dcc8c8c4c65
  - user:07c4158f-152b-4071-a1d9-92b8b56e4a45
  ...
```

### 2. 커스텀 크기

```bash
# 20만 사용자, 10만 문서, 5만 폴더
go run scripts/generate_testdata.go -users 200000 -docs 100000 -folders 50000
```

### 3. 프로그래밍 방식

```go
import "github.com/carped99/steampipe-plugin-openfga/openfga"

config := openfga.TestDataConfig{
    NumUsers:   100000,
    NumDocs:    50000,
    NumFolders: 20000,
    BatchSize:  1000,
}

data, err := openfga.GenerateTestData(config)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Generated %d users\n", len(data.Users))
fmt.Printf("Generated %d docs\n", len(data.Docs))
fmt.Printf("Generated %d folders\n", len(data.Folders))
```

## 📁 생성되는 데이터 형식

### Users
```
user:d26cb748-2624-4af4-98d7-2dcc8c8c4c65
user:07c4158f-152b-4071-a1d9-92b8b56e4a45
user:34d4785b-d3f3-4245-9866-aca8968a8ca0
```

### Docs
```
doc:dbcc0202-910e-4ec9-92b1-80da72f23a67
doc:53a2c07f-5793-4822-af24-07b7be10a96a
doc:e205cf36-7721-4d70-9d80-4938c6cae9da
```

### Folders
```
folder:657e52af-8d1b-46b0-86ff-c5a22dcca2cd
folder:9ac284a1-a9fb-42cb-b4bf-5555fa2a9e0c
folder:dfc73c2a-c778-4630-a4de-3d38d6fd67a6
```

## 🧪 테스트

### 기본 테스트
```bash
go test ./openfga/... -v -run TestGenerateTestData
```

### 대용량 테스트 (10만+)
```bash
go test ./openfga/... -v -run TestGenerateLargeTestData
```

### 벤치마크
```bash
go test ./openfga/... -bench=BenchmarkTestDataGeneration
```

## 📦 의존성

```go
require github.com/go-faker/faker/v4 v4.7.0
```

## 💡 구현 세부사항

### 데이터 구조

```go
type FakeUser struct {
    ID       string `faker:"uuid_hyphenated"`
    Username string `faker:"username"`
}

type FakeDoc struct {
    ID    string `faker:"uuid_hyphenated"`
    Title string `faker:"sentence"`
}

type FakeFolder struct {
    ID   string `faker:"uuid_hyphenated"`
    Name string `faker:"word"`
}
```

### 병렬 처리

```go
var wg sync.WaitGroup

// Users 생성
wg.Add(1)
go func() {
    defer wg.Done()
    users := generateUsers(config.NumUsers)
    // ...
}()

// Docs 생성
wg.Add(1)
go func() {
    defer wg.Done()
    docs := generateDocs(config.NumDocs)
    // ...
}()

// Folders 생성
wg.Add(1)
go func() {
    defer wg.Done()
    folders := generateFolders(config.NumFolders)
    // ...
}()

wg.Wait()
```

## 🎯 향후 계획

- [ ] OpenFGA Write API를 통한 튜플 자동 생성
- [ ] Batch Write 지원으로 대량 삽입 성능 개선
- [ ] 다양한 relation 패턴 지원 (viewer, editor, owner 등)
- [ ] CSV/JSON 파일 export 기능
- [ ] 데이터 분포 커스터마이징 (Zipf distribution 등)

## 📝 예제 출력

```
=== OpenFGA Test Data Generator ===
Configuration:
  Users: 100000
  Docs: 50000
  Folders: 20000

Generated 20000 folders
Generated 50000 docs
Generated 100000 users
Test data generation completed in 753.1677ms
Total entities: 170000 (users: 100000, docs: 50000, folders: 20000)

=== Generation Report ===
Total entities: 170000
  - Users: 100000
  - Docs: 50000
  - Folders: 20000
Generation time: 753.1677ms
Entities per second: 225713.34

Sample Users (first 5):
  - user:d26cb748-2624-4af4-98d7-2dcc8c8c4c65
  - user:07c4158f-152b-4071-a1d9-92b8b56e4a45
  - user:34d4785b-d3f3-4245-9866-aca8968a8ca0
  - user:0e6bf034-5414-4a11-9b04-4ea1ef7d4d42
  - user:d7a4e5bb-f987-4a6e-b17b-b3fd5c7673ed

Sample Docs (first 5):
  - doc:dbcc0202-910e-4ec9-92b1-80da72f23a67
  - doc:53a2c07f-5793-4822-af24-07b7be10a96a
  - doc:e205cf36-7721-4d70-9d80-4938c6cae9da
  - doc:3ae84ef7-5c5f-4e06-86b8-f0d716913d78
  - doc:91262295-160f-4412-8499-fb4b5019913d

Sample Folders (first 5):
  - folder:657e52af-8d1b-46b0-86ff-c5a22dcca2cd
  - folder:9ac284a1-a9fb-42cb-b4bf-5555fa2a9e0c
  - folder:dfc73c2a-c778-4630-a4de-3d38d6fd67a6
  - folder:ca52edda-1ce7-40cd-a267-a121bb05fd24
  - folder:5aabc63d-97d5-497c-b39f-50e97013214f

=== Completed Successfully ===
```

## 🔧 트러블슈팅

### 메모리 부족
대량 데이터 생성 시 메모리가 부족할 경우, BatchSize를 늘려서 여러 번에 나눠 생성하세요.

### 성능 저하
- CPU 코어 수 확인: `runtime.NumCPU()`
- GOMAXPROCS 설정 확인
- 가비지 컬렉터 튜닝 고려

## 📄 라이선스

이 프로젝트의 라이선스를 따릅니다.
