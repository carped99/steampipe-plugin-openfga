package openfga

import (
	"context"
	"fmt"
	openfgav1 "github.com/carped99/steampipe-plugin-openfga/internal/openfga/gen/openfga/v1"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"strings"
	"time"
)

type AclPermissionRow struct {
	ObjectType  string `json:"object_type"`
	ObjectId    string `json:"object_id"`
	SubjectType string `json:"subject_type"`
	SubjectId   string `json:"subject_id"`
	Relation    string
	CheckedAt   time.Time
}

func tableAclPermission() *plugin.Table {
	return &plugin.Table{
		Name:        "sys_acl_permission",
		Description: "Real-time permission check via OpenFGA Check API",
		List: &plugin.ListConfig{
			Hydrate: checkPermission,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "object_type", Require: plugin.Optional},
				{Name: "object_id", Require: plugin.Optional},
				{Name: "subject_type", Require: plugin.Optional},
				{Name: "subject_id", Require: plugin.Optional},
				{Name: "relation", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "object_type", Type: proto.ColumnType_STRING, Description: "Logical type of the protected object"},
			{Name: "object_id", Type: proto.ColumnType_STRING, Description: "Application-level identifier of the object"},
			{Name: "subject_type", Type: proto.ColumnType_STRING, Description: "Type of the subject (e.g. 'user', 'group', 'service')"},
			{Name: "subject_id", Type: proto.ColumnType_STRING, Description: "Identifier of the subject (user ID, group ID etc)"},
			{Name: "relation", Type: proto.ColumnType_STRING, Description: "Relation to check(e.g. 'reader', 'writer')"},
			{Name: "allowed", Type: proto.ColumnType_BOOL, Description: "Whether the permission is granted"},
			{Name: "policy_version", Type: proto.ColumnType_STRING, Description: "Authorization model or snapshot version used to evaluate this permission."},
			{Name: "evaluated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when this effective permission was evaluated"},
		},
	}
}

// checkPermission Steampipe List 요청을 받아서 OpenFGA 쪽으로 위임하고 (subject, object, relation) 단위로 효과적인 권한을 스트리밍.
func checkPermission(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (any, error) {
	// Extract query parameters first to validate before connecting
	objectType := d.EqualsQuals["object_type"].GetStringValue()
	objectId := d.EqualsQuals["object_id"].GetStringValue()
	subjectType := d.EqualsQuals["subject_type"].GetStringValue()
	subjectId := d.EqualsQuals["subject_id"].GetStringValue()

	relation := ""
	if v, ok := d.EqualsQuals["relation"]; ok {
		relation = v.GetStringValue()
	}

	// Early return if required fields are missing
	if subjectType == "" || subjectId == "" {
		return nil, nil
	}

	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}

	user := fmt.Sprintf("%s:%s", subjectType, subjectId)
	object := fmt.Sprintf("%s:%s", objectType, objectId)

	var continuationToken string
	for {
		rowsRemaining := d.RowsRemaining(ctx)
		if rowsRemaining == 0 {
			break
		}

		pageSize := int32(rowsRemaining)
		if pageSize > 100 {
			pageSize = 100
		}

		request := &openfgav1.ReadRequest{
			StoreId: client.storeID,
			TupleKey: &openfgav1.ReadRequestTupleKey{
				User:     user,
				Object:   object,
				Relation: relation,
			},
			PageSize:          wrapperspb.Int32(pageSize),
			ContinuationToken: continuationToken,
		}

		response, err := client.Read(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, t := range response.Tuples {
			if t.Key == nil {
				continue
			}

			d.StreamListItem(ctx, parseAclPermissionRow(t.Key))
		}

		if response.ContinuationToken == "" {
			break
		}
		continuationToken = response.ContinuationToken
	}

	return nil, nil
}

func parseAclPermissionRow(key *openfgav1.TupleKey) AclPermissionRow {
	objectType, objectId := splitObject(key.Object)
	subjectType, subjectId := splitObject(key.User)
	return AclPermissionRow{
		ObjectType:  objectType,
		ObjectId:    objectId,
		SubjectType: subjectType,
		SubjectId:   subjectId,
		Relation:    key.Relation,
	}
}

func splitObject(obj string) (objectType, objectID string) {
	parts := strings.SplitN(obj, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return obj, ""
}
