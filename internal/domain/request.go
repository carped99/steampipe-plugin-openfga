package domain

type CheckRequest struct {
	objectType  string
	objectId    string
	subjectType string
	subjectId   string
	relation    string
}
