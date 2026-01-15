package shared

type Client interface {
	Close() error
	ListDirectory(subdirectory string) (Items, error)
	BuildDownloadHeaders() map[string]string
}
