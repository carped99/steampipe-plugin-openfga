package openfga

import (
	"context"
	"fmt"
	"github.com/openfga/go-sdk/client"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"time"
)

func tableAclPermission() *plugin.Table {
	return &plugin.Table{
		Name:        "sys_acl_permission",
		Description: "Real-time permission check via OpenFGA Check API",
		List: &plugin.ListConfig{
			Hydrate: checkPermission,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "object_type", Require: plugin.Required},
				{Name: "object_id", Require: plugin.Required},
				{Name: "subject_type", Require: plugin.Required},
				{Name: "subject_id", Require: plugin.Required},
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

type CheckPermissionRow struct {
	UserId     string
	Relation   string
	ObjectType string
	ObjectId   string
	Allowed    bool
	CheckedAt  time.Time
}

// checkPermission Steampipe List 요청을 받아서 OpenFGA 쪽으로 위임하고 (subject, object, relation) 단위로 효과적인 권한을 스트리밍.
func checkPermission(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (any, error) {
	fgaClient, err := connect(ctx, d)
	if err != nil {
		return nil, err
	}

	objectType := d.EqualsQuals["object_type"].GetStringValue()
	objectId := d.EqualsQuals["object_id"].GetStringValue()
	subjectType := d.EqualsQuals["subject_type"].GetStringValue()
	subjectId := d.EqualsQuals["subject_id"].GetStringValue()

	relation := ""
	if v, ok := d.EqualsQuals["relation"]; ok {
		relation = v.GetStringValue()
	}

	if subjectType == "" || subjectId == "" {
		return nil, nil
	}

	user := fmt.Sprintf("%s:%s", subjectType, subjectId)
	object := fmt.Sprintf("%s:%s", objectType, objectId)

	// OpenFGA Check API 호출
	checkRequest := client.ClientCheckRequest{
		User:     user,
		Relation: relation,
		Object:   object,
	}

	response, err := fgaClient.Check(ctx).Body(checkRequest).Execute()
	if err != nil {
		return nil, err
	}

	result := &CheckPermissionRow{
		UserId:     subjectId,
		Relation:   relation,
		ObjectType: objectType,
		ObjectId:   objectId,
		Allowed:    response.GetAllowed(),
		CheckedAt:  time.Now(),
	}

	d.StreamListItem(ctx, result)
	return nil, nil
}
