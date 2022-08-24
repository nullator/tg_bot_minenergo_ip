package databases

type Database interface {
	Save(key string, value string, bucket string) error
	Get(key string, bucket string) (string, error)
	GetAll(bucket string) (map[string]string, error)
}
