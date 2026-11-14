package usecases

type Object interface {
	Get(key string) (string, error)
	Post(key string, value string) (string, error)
	Put(key string, value string) (string, error)
	Delete(key string) error
}
