package db

type IMorphTabler interface {
	PHPackage() string
	GetID() int64
}
