package domain

type AclService interface {
	Check(request CheckRequest) (bool, error)
	BatchCheck(requests []CheckRequest) (bool, error)
}
