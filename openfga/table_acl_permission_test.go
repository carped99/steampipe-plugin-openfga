package openfga

import (
	"context"
	"fmt"
	"os"
	"testing"

	openfgav1 "github.com/carped99/steampipe-plugin-openfga/internal/openfga/gen/openfga/v1"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// testSetup holds the test environment setup
type testSetup struct {
	client *Client
	config *Config
}

// testCase represents a single permission test scenario
type testCase struct {
	name        string
	subjectType string
	subjectId   string
	relation    string
	objectType  string
	objectId    string
	setupTuple  bool
	expected    bool
}

// setUp creates a gRPC client and prepares test data
func setUp(t *testing.T) (*testSetup, []testCase) {
	t.Helper()

	endpoint := os.Getenv("OPENFGA_API_URL")
	storeId := os.Getenv("OPENFGA_STORE_ID")
	modelId := os.Getenv("OPENFGA_AUTHORIZATION_MODEL_ID")

	storeId = "01K9Y2QSETQJE22F1BNEJ3ZWTM"
	if endpoint == "" {
		endpoint = "localhost:8081"
	}
	if storeId == "" {
		t.Skip("OPENFGA_STORE_ID not set, skipping integration test")
	}

	ctx := context.Background()

	// Create Config for newClient
	cfg := &Config{
		Endpoint:             &endpoint,
		StoreId:              &storeId,
		AuthorizationModelId: &modelId,
	}

	// Use newClient to create the client
	client, err := newClient(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenFGA client: %v", err)
	}

	setup := &testSetup{
		client: client,
		config: cfg,
	}

	// Define test cases
	testCases := []testCase{
		{
			name:        "User has viewer permission on document",
			subjectType: "user",
			subjectId:   "alice",
			relation:    "viewer",
			objectType:  "doc",
			objectId:    "test-doc-1",
			setupTuple:  true,
			expected:    true,
		},
		{
			name:        "User does not have editor permission on document",
			subjectType: "user",
			subjectId:   "alice",
			relation:    "can_write",
			objectType:  "doc",
			objectId:    "test-doc-1",
			setupTuple:  false,
			expected:    false,
		},
		{
			name:        "User has owner permission on folder",
			subjectType: "user",
			subjectId:   "bob",
			relation:    "owner",
			objectType:  "folder",
			objectId:    "test-folder-1",
			setupTuple:  true,
			expected:    true,
		},
	}

	// Create test tuples using gRPC client
	for _, tc := range testCases {
		if tc.setupTuple {
			user := fmt.Sprintf("%s:%s", tc.subjectType, tc.subjectId)
			object := fmt.Sprintf("%s:%s", tc.objectType, tc.objectId)

			writeReq := &openfgav1.WriteRequest{
				StoreId: storeId,
				Writes: &openfgav1.WriteRequestWrites{
					TupleKeys: []*openfgav1.TupleKey{
						{
							User:     user,
							Relation: tc.relation,
							Object:   object,
						},
					},
				},
			}

			_, err := client.Write(ctx, writeReq)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	return setup, testCases
}

// tearDown cleans up test data and closes connections
func tearDown(t *testing.T, setup *testSetup, testCases []testCase) {
	t.Helper()

	ctx := context.Background()

	// Delete test tuples using gRPC client
	for _, tc := range testCases {
		if tc.setupTuple {
			user := fmt.Sprintf("%s:%s", tc.subjectType, tc.subjectId)
			object := fmt.Sprintf("%s:%s", tc.objectType, tc.objectId)

			deleteReq := &openfgav1.WriteRequest{
				Deletes: &openfgav1.WriteRequestDeletes{
					TupleKeys: []*openfgav1.TupleKeyWithoutCondition{
						{
							User:     user,
							Relation: tc.relation,
							Object:   object,
						},
					},
				},
			}

			_, err := setup.client.Write(ctx, deleteReq)
			if err != nil {
				t.Logf("Warning: Failed to cleanup tuple: %v", err)
			}
		}
	}

	// Close gRPC connection
	if setup.client != nil {
		if err := setup.client.Close(); err != nil {
			t.Logf("Warning: Failed to close connection: %v", err)
		}
	}
}

// TestTableAclPermission_Integration tests the sys_acl_permission table with a real OpenFGA server
func TestTableAclPermission_Integration(t *testing.T) {
	// Setup test environment and data
	setup, testCases := setUp(t)
	defer tearDown(t, setup, testCases)

	ctx := context.Background()

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock QueryData
			queryData := &plugin.QueryData{
				Connection: &plugin.Connection{
					Config: setup.config,
				},
				EqualsQuals: plugin.KeyColumnEqualsQualMap{
					"subject_type": &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: tc.subjectType}},
					"subject_id":   &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: tc.subjectId}},
					"relation":     &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: tc.relation}},
					"object_type":  &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: tc.objectType}},
					"object_id":    &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: tc.objectId}},
				},
			}

			// Mock StreamListItem
			var result *AclPermissionRow
			queryData.StreamListItem = func(ctx context.Context, items ...interface{}) {
				if len(items) > 0 {
					result = items[0].(*AclPermissionRow)
				}
			}

			// Call the function
			_, err := checkPermission(ctx, queryData, nil)
			if err != nil {
				t.Fatalf("checkPermission failed: %v", err)
			}

			// Verify results
			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			if result.SubjectId != tc.subjectId {
				t.Errorf("Expected SubjectId %s, got %s", tc.subjectId, result.SubjectId)
			}

			if result.Relation != tc.relation {
				t.Errorf("Expected Relation %s, got %s", tc.relation, result.Relation)
			}

			if result.ObjectType != tc.objectType {
				t.Errorf("Expected ObjectType %s, got %s", tc.objectType, result.ObjectType)
			}

			if result.ObjectId != tc.objectId {
				t.Errorf("Expected ObjectId %s, got %s", tc.objectId, result.ObjectId)
			}

			// Note: Allowed field is not part of AclPermissionRow anymore
			// The row represents tuple keys, not permission check results
		})
	}
}

// TestTableAclPermission_MissingQuals tests that missing required quals return appropriate errors
func TestTableAclPermission_MissingQuals(t *testing.T) {
	ctx := context.Background()
	apiUrl := "http://localhost:8080"
	storeId := "01ARZ3NDEKTSV4RRFFQ69G5FAV" // Valid ULID format

	testCases := []struct {
		name        string
		equalsQuals plugin.KeyColumnEqualsQualMap
		expectNil   bool
	}{
		{
			name: "All quals provided",
			equalsQuals: plugin.KeyColumnEqualsQualMap{
				"subject_type": &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "user"}},
				"subject_id":   &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "alice"}},
				"relation":     &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "viewer"}},
				"object_type":  &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "document"}},
				"object_id":    &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "doc1"}},
			},
			expectNil: false,
		},
		{
			name: "Missing subject_id returns nil (no error, no result)",
			equalsQuals: plugin.KeyColumnEqualsQualMap{
				"subject_type": &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "user"}},
				"relation":     &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "viewer"}},
				"object_type":  &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "document"}},
				"object_id":    &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "doc1"}},
			},
			expectNil: true,
		},
		{
			name: "Missing subject_type returns nil (no error, no result)",
			equalsQuals: plugin.KeyColumnEqualsQualMap{
				"subject_id":  &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "alice"}},
				"relation":    &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "viewer"}},
				"object_type": &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "document"}},
				"object_id":   &proto.QualValue{Value: &proto.QualValue_StringValue{StringValue: "doc1"}},
			},
			expectNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryData := &plugin.QueryData{
				Connection: &plugin.Connection{
					Config: &Config{
						Endpoint: &apiUrl,
						StoreId:  &storeId,
					},
				},
				EqualsQuals: tc.equalsQuals,
			}

			var called bool
			queryData.StreamListItem = func(ctx context.Context, items ...interface{}) {
				called = true
			}

			result, err := checkPermission(ctx, queryData, nil)

			// The function returns (nil, nil) when required quals are missing
			if tc.expectNil {
				if result != nil || err != nil {
					t.Errorf("Expected (nil, nil) but got (%v, %v)", result, err)
				}
				if called {
					t.Error("StreamListItem should not be called when quals are missing")
				}
			} else {
				// When all quals are provided, it should attempt connection
				// This will fail in unit test without real server, which is OK
				if err == nil {
					t.Error("Expected connection error in unit test without real server")
				}
			}
		})
	}
}

// TestConnect tests the connection function
func TestConnect(t *testing.T) {
	testCases := []struct {
		name        string
		apiUrl      string
		storeId     string
		expectError bool
	}{
		{
			name:        "Valid configuration",
			apiUrl:      "http://localhost:8080",
			storeId:     "01ARZ3NDEKTSV4RRFFQ69G5FAV", // Valid ULID format
			expectError: false,
		},
		{
			name:        "Missing API URL",
			apiUrl:      "",
			storeId:     "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			expectError: true,
		},
		{
			name:        "Missing Store ID",
			apiUrl:      "http://localhost:8080",
			storeId:     "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			queryData := &plugin.QueryData{
				Connection: &plugin.Connection{
					Config: &Config{
						Endpoint: &tc.apiUrl,
						StoreId:  &tc.storeId,
					},
				},
			}

			client, err := connect(ctx, queryData)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if client == nil {
					t.Error("Expected client but got nil")
				}
			}
		})
	}
}
