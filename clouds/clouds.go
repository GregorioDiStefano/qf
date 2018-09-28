package clouds

import "io"

type Cloud interface {
	ListObjects() ([]string, error)
	ListObjectsWithPrefix(string) ([]string, error)
	//GetMetadata(string) (map[string]string, error)
	//SetMetadata(string, map[string]string) error

	Download(string) (io.Reader, error)
	Upload(string, io.Reader) error
	Remove(string) error
}
