package persistent

type Storage interface {
	Save() error
	Restore() error
}
