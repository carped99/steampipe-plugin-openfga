package openfga

import (
	"context"
	"errors"
	"fmt"
	openfgav1 "github.com/carped99/steampipe-plugin-openfga/internal/openfga/gen/openfga/v1"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"io"
	"time"
)

// (object_type, object_id, subject_type, subject_id, relation)
//	→ Check 1회
//
//(object_type, subject_type, subject_id, relation)
//→ “이 subject 가 이 relation 으로 접근 가능한 object_type 모든 object”
//→ ListObjects
//
//(object_type, object_id, relation)
//→ “이 object 에 이 relation 을 가진 모든 subject”
//→ ListUsers
//
//(subject_type, subject_id, relation) (object_type 없이)
//→ “이 subject 가 이 relation 을 가진 모든 object (모든 type)”
//→ Read (user+relation filter) + 클라이언트에서 type:id 파싱

type AclPermissionRow struct {
	ObjectType  string    `json:"object_type"`
	ObjectID    string    `json:"object_id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	Relation    string    `json:"relation"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

var (
	objectTypeCol  = "object_type"
	objectIDCol    = "object_id"
	subjectTypeCol = "subject_type"
	subjectIDCol   = "subject_id"
	relationCol    = "relation"
)

func tableAclPermission(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "sys_acl_permission",
		Description: "Real-time permission check via OpenFGA Check API",
		List: &plugin.ListConfig{
			Hydrate: listPermission,
			KeyColumns: []*plugin.KeyColumn{
				{Name: relationCol, Require: plugin.Required},
			},
		},
		Columns: []*plugin.Column{
			{Name: objectTypeCol, Type: proto.ColumnType_STRING, Description: "Logical type of the protected object"},
			{Name: objectIDCol, Type: proto.ColumnType_STRING, Description: "Application-level identifier of the object"},
			{Name: subjectTypeCol, Type: proto.ColumnType_STRING, Description: "Type of the subject (e.g. 'user', 'group', 'service')"},
			{Name: subjectIDCol, Type: proto.ColumnType_STRING, Description: "Identifier of the subject (user ID, group ID etc)"},
			{Name: relationCol, Type: proto.ColumnType_STRING, Description: "Relation to check(e.g. 'reader', 'writer')"},
			{Name: "policy_version", Type: proto.ColumnType_STRING, Description: "Authorization model or snapshot version used to evaluate this permission."},
			{Name: "evaluated_at", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp when this effective permission was evaluated"},
		},
	}
}

// listPermission Steampipe List 요청을 받아서 OpenFGA 쪽으로 위임하고 (subject, object, relation) 단위로 효과적인 권한을 스트리밍.
func listPermission(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (any, error) {
	logger := plugin.Logger(ctx)
	if logger.IsDebug() {
		logger.Debug("listPermission called", "quals", d.EqualsQuals)
	}

	objectType := d.EqualsQualString(objectTypeCol)
	objectId := d.EqualsQualString(objectIDCol)
	subjectType := d.EqualsQualString(subjectTypeCol)
	subjectId := d.EqualsQualString(subjectIDCol)
	relation := d.EqualsQualString(relationCol)

	if relation == "" {
		return nil, fmt.Errorf("relation is required")
	}

	hasObject := objectType != "" && objectId != ""
	hasSubject := subjectType != "" && subjectId != ""
	switch {
	// 1) 단건 조회 - 모든 정보가 있을 때
	case hasObject && hasSubject:
		return check(ctx, d, objectType, objectId, subjectType, subjectId, relation)
	// 2) subject 있음, object_id 값이 없음
	case hasSubject && objectType != "":
		return listObjects(ctx, d, objectType, subjectType, subjectId, relation)
	// 3) object 있음, subject_id 값이 없음
	case hasObject && subjectType != "":
		return listUsers(ctx, d, objectType, objectId, subjectType, relation)
	default:
		return listByRead(ctx, d, objectType, objectId, subjectType, subjectId, relation)
	}
}

func listObjects(ctx context.Context, d *plugin.QueryData, objectType, subjectType, subjectID, relation string) (any, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}

	req := &openfgav1.StreamedListObjectsRequest{
		StoreId:  client.storeID,
		Relation: relation,
		User:     subjectType + ":" + subjectID,
		Type:     objectType,
	}

	res, err := client.StreamedListObjects(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenFGA ListObjects: %w", err)
	}

	for {
		chunk, err := res.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		prefix, id := splitObject(chunk.Object)
		row := AclPermissionRow{
			ObjectType:  prefix,
			ObjectID:    id,
			SubjectType: subjectType,
			SubjectID:   subjectID,
			Relation:    relation,
		}
		d.StreamListItem(ctx, row)

		if d.RowsRemaining(ctx) == 0 {
			break
		}
	}

	return nil, nil
}

func listUsers(ctx context.Context, d *plugin.QueryData, objectType, objectID, subjectType, relation string) (any, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}

	req := &openfgav1.ListUsersRequest{
		StoreId:  client.storeID,
		Relation: relation,
		Object: &openfgav1.Object{
			Type: objectType,
			Id:   objectID,
		},
		UserFilters: []*openfgav1.UserTypeFilter{
			{
				Type: subjectType,
			},
		},
		Consistency: openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY,
	}

	res, err := client.ListUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenFGA ListUsers: %w", err)
	}

	t := time.Now().UTC()
	for _, user := range res.GetUsers() {
		switch v := user.GetUser().(type) {
		case *openfgav1.User_Object:
			obj := v.Object
			if obj == nil {
				continue
			}
			row := AclPermissionRow{
				ObjectType:  objectType,
				ObjectID:    objectID,
				SubjectType: obj.Type,
				SubjectID:   obj.Id,
				Relation:    relation,
				EvaluatedAt: t,
			}
			d.StreamListItem(ctx, row)

		case *openfgav1.User_Userset:
			return nil, fmt.Errorf("userset subjects are not supported in ListUsers results")
			//us := v.Userset
			//if us == nil {
			//	continue
			//}
			//out = append(out, ACLSubject{
			//	Kind:        SubjectKindUserset,
			//	SubjectType: us.GetType(),     // "group"
			//	SubjectID:   us.GetId(),       // "eng"
			//	UsersetRel:  us.GetRelation(), // "member"
			//})

		case *openfgav1.User_Wildcard:
			return nil, fmt.Errorf("wildcard subjects are not supported in ListUsers results")
			//w := v.Wildcard
			//if w == nil {
			//	continue
			//}
			//out = append(out, ACLSubject{
			//	Kind:        SubjectKindWildcard,
			//	SubjectType: w.GetType(), // "user"
			//	SubjectID:   "*",         // 관례적으로, 전체
			//})
		}
	}
	return nil, nil
}

func check(ctx context.Context, d *plugin.QueryData, objectType, objectId, subjectType, subjectId, relation string) (any, error) {
	logger := plugin.Logger(ctx)
	if logger.IsDebug() {
		logger.Debug("check called", "quals", d.EqualsQuals)
	}

	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}

	req := &openfgav1.CheckRequest{
		StoreId: client.storeID,
		TupleKey: &openfgav1.CheckRequestTupleKey{
			Object:   objectType + ":" + objectId,
			User:     subjectType + ":" + subjectId,
			Relation: relation,
		},
		Consistency: openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY,
	}

	res, err := client.Check(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenFGA Check: %w", err)
	}

	if res.GetAllowed() {
		row := AclPermissionRow{
			ObjectType:  objectType,
			ObjectID:    objectId,
			SubjectType: subjectType,
			SubjectID:   subjectId,
			Relation:    relation,
			EvaluatedAt: time.Now().UTC(),
		}
		d.StreamListItem(ctx, row)
	}

	return nil, nil
}

// listByRead does not evaluate the tuples in the store.
//  1. tuple_key is optional. If not specified, it will return all tuples in the store.
//  2. tuple_key.object is mandatory if tuple_key is specified.
//     It can be a full object (e.g., type:object_id) or type only (e.g., type:)
//  3. tuple_key.user is mandatory if tuple_key is specified in the case the tuple_key.object is a type only.
//     If tuple_key.user is specified, it needs to be a full object (e.g., type:user_id).
func listByRead(ctx context.Context, d *plugin.QueryData, objectType, objectId, subjectType, subjectId, relation string) (any, error) {
	client, err := getClient(ctx, d)
	if err != nil {
		return nil, err
	}

	var tupleKey *openfgav1.ReadRequestTupleKey

	if objectType != "" {
		tupleKey = &openfgav1.ReadRequestTupleKey{
			Object: objectType + ":" + objectId,
		}

		if objectId == "" {
			// object_id 가 없으면 -> user 필수
			if subjectType == "" || subjectId == "" {
				return nil, fmt.Errorf("subject_type and subject_id are required when object_id is not specified")
			}
			tupleKey.User = subjectType + ":" + subjectId
		}
		tupleKey.Relation = relation
	}

	var continuationToken string
	for {
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}

		req := &openfgav1.ReadRequest{
			StoreId:           client.storeID,
			TupleKey:          tupleKey,
			ContinuationToken: continuationToken,
			//PageSize:          wrapperspb.Int32(100), // Use default page size
			Consistency: openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY,
		}

		res, err := client.Read(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("OpenFGA Read: %w", err)
		}

		// ReadResponse.Tuples: []*Tuple
		for _, t := range res.GetTuples() {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}

			key := t.GetKey()
			if key == nil {
				continue
			}

			key.GetObject()
			key.GetUser()

			listObjects(ctx, d, objectType, subjectType, subjectId, relation)

			// key.Object 는 "doc:123" 이런 문자열
			objType, objID := splitObject(key.GetObject())

			// key.User 는 "user:alice" 같은 문자열
			subjType, subjID := splitObject(key.GetUser())

			row := AclPermissionRow{
				ObjectType:  objType,
				ObjectID:    objID,
				SubjectType: subjType,
				SubjectID:   subjID,
				Relation:    key.GetRelation(),
				EvaluatedAt: t.Timestamp.AsTime(),
			}

			d.StreamListItem(ctx, row)
		}

		continuationToken = res.GetContinuationToken()
		if continuationToken == "" {
			break
		}
	}
	return nil, nil
}
