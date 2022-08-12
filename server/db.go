package server

type KV interface {
	Bucket() string
	Key() string
	Value() []byte
	Encrypt() error
}

type Watcher interface {
	Watch()
}

type Backend interface {
	Watcher
}
