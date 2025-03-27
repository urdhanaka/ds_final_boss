package db

type DatabaseInterface interface {
	Store(nodeModel NodesModel) error
	Delete(nodeModel NodesModel) error
}
