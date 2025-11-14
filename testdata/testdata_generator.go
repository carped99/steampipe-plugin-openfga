package main

import (
	"context"
	"fmt"
	openfgav1 "github.com/carped99/steampipe-plugin-openfga/internal/openfga/gen/openfga/v1"
	"github.com/carped99/steampipe-plugin-openfga/openfga"
	"github.com/go-faker/faker/v4"
	"sync"
	"time"
)

// GeneratedTestData holds generated test entities
type GeneratedTestData struct {
	Users          []string
	Docs           []string
	Folders        []string
	GenerationTime time.Duration
}

// FakeUser generates a fake user ID
type FakeUser struct {
	ID       string `faker:"uuid_hyphenated"`
	Username string `faker:"username"`
}

// FakeDoc generates a fake document
type FakeDoc struct {
	ID    string `faker:"uuid_hyphenated"`
	Title string `faker:"sentence"`
}

// FakeFolder generates a fake folder
type FakeFolder struct {
	ID   string `faker:"uuid_hyphenated"`
	Name string `faker:"word"`
}

// generateTestData generates large amounts of test data using faker
func generateTestData(cfg config) (*GeneratedTestData, error) {
	start := time.Now()

	data := &GeneratedTestData{
		Users:   make([]string, 0, cfg.Generation.Users),
		Docs:    make([]string, 0, cfg.Generation.Docs),
		Folders: make([]string, 0, cfg.Generation.Folders),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Generate users concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		users := generateUsers(cfg.Generation.Users)
		mu.Lock()
		data.Users = users
		mu.Unlock()
		fmt.Printf("Generated %d users\n", len(users))
	}()

	// Generate docs concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		docs := generateDocs(cfg.Generation.Docs)
		mu.Lock()
		data.Docs = docs
		mu.Unlock()
		fmt.Printf("Generated %d docs\n", len(docs))
	}()

	// Generate folders concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		folders := generateFolders(cfg.Generation.Folders)
		mu.Lock()
		data.Folders = folders
		mu.Unlock()
		fmt.Printf("Generated %d folders\n", len(folders))
	}()

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Test data generation completed in %s\n", elapsed)
	fmt.Printf("Total entities: %d (users: %d, docs: %d, folders: %d)\n",
		len(data.Users)+len(data.Docs)+len(data.Folders),
		len(data.Users), len(data.Docs), len(data.Folders))

	return data, nil
}

func generateUsers(count int) []string {
	users := make([]string, count)
	for i := 0; i < count; i++ {
		var user FakeUser
		_ = faker.FakeData(&user)
		users[i] = user.ID
	}
	return users
}

func generateDocs(count int) []string {
	docs := make([]string, count)
	for i := 0; i < count; i++ {
		var doc FakeDoc
		_ = faker.FakeData(&doc)
		docs[i] = doc.ID
	}
	return docs
}

func generateFolders(count int) []string {
	folders := make([]string, count)
	for i := 0; i < count; i++ {
		var folder FakeFolder
		_ = faker.FakeData(&folder)
		folders[i] = folder.ID
	}
	return folders
}

// CreateTuplesFromTestData creates OpenFGA tuples from generated test data
func CreateTuplesFromTestData(ctx context.Context, client *openfga.Client, storeID string, data *GeneratedTestData, relations []string) error {
	const batchSize = 100
	start := time.Now()
	totalTuples := 0

	batch := make([]tupleSpec, 0, batchSize)

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := createTuples(ctx, client, storeID, batch); err != nil {
			return err
		}
		totalTuples += len(batch)
		batch = batch[:0]
		return nil
	}

	// Create user->doc relations
	fmt.Println("\nCreating user->doc tuples...")
	for i, user := range data.Users {
		// Each user gets access to 3-5 docs
		numDocs := 3 + (i % 3)
		for j := 0; j < numDocs && j < len(data.Docs); j++ {
			docIdx := (i + j) % len(data.Docs)
			relation := relations[j%len(relations)]

			batch = append(batch, tupleSpec{
				subject:  fmt.Sprintf("user:%s", user),
				relation: relation,
				object:   fmt.Sprintf("doc:%s", data.Docs[docIdx]),
			})

			if len(batch) >= batchSize {
				if err := flushBatch(); err != nil {
					return fmt.Errorf("failed to create user->doc tuples: %w", err)
				}
				fmt.Printf("  Progress: %d tuples created\n", totalTuples)
			}
		}
	}

	// Flush remaining doc tuples
	if err := flushBatch(); err != nil {
		return fmt.Errorf("failed to create user->doc tuples: %w", err)
	}
	docTuples := totalTuples
	fmt.Printf("  Completed: %d doc tuples created\n", docTuples)

	// Create user->folder relations
	fmt.Println("\nCreating user->folder tuples...")
	for i, user := range data.Users {
		// Each user gets access to 2-3 folders
		numFolders := 2 + (i % 2)
		for j := 0; j < numFolders && j < len(data.Folders); j++ {
			folderIdx := (i + j) % len(data.Folders)
			relation := relations[j%len(relations)]

			batch = append(batch, tupleSpec{
				subject:  fmt.Sprintf("user:%s", user),
				relation: relation,
				object:   fmt.Sprintf("folder:%s", data.Folders[folderIdx]),
			})

			if len(batch) >= batchSize {
				if err := flushBatch(); err != nil {
					return fmt.Errorf("failed to create user->folder tuples: %w", err)
				}
				fmt.Printf("  Progress: %d tuples created\n", totalTuples)
			}
		}
	}

	// Flush remaining folder tuples
	if err := flushBatch(); err != nil {
		return fmt.Errorf("failed to create user->folder tuples: %w", err)
	}
	folderTuples := totalTuples - docTuples
	fmt.Printf("  Completed: %d folder tuples created\n", folderTuples)

	elapsed := time.Since(start)
	fmt.Printf("\nTuple creation completed in %s\n", elapsed)
	fmt.Printf("Total tuples created: %d\n", totalTuples)

	return nil
}

func createTuples(ctx context.Context, client *openfga.Client, storeID string, tuples []tupleSpec) error {
	if len(tuples) == 0 {
		return nil
	}

	tupleKeys := make([]*openfgav1.TupleKey, 0, len(tuples))
	for _, t := range tuples {
		objectType, objectID := splitObject(t.object)
		subjectType, subjectID := splitObject(t.subject)

		tupleKey := &openfgav1.TupleKey{
			Object:   fmt.Sprintf("%s:%s", objectType, objectID),
			Relation: t.relation,
			User:     fmt.Sprintf("%s:%s", subjectType, subjectID),
		}
		tupleKeys = append(tupleKeys, tupleKey)
	}

	writeRequest := &openfgav1.WriteRequest{
		StoreId: storeID,
		Writes: &openfgav1.WriteRequestWrites{
			TupleKeys: tupleKeys,
		},
	}

	_, err := client.Write(ctx, writeRequest)
	return err
}

type tupleSpec struct {
	subject  string
	relation string
	object   string
}

func splitObject(obj string) (objectType, objectID string) {
	for i, c := range obj {
		if c == ':' {
			return obj[:i], obj[i+1:]
		}
	}
	return obj, ""
}
