package main

type ConfigLoader interface {
	Load(string) error
	GetTgts() []string
}

type HandlerConfigLoader interface {
	Load(string) error
	GetBackupDir() string
	GetHistoryCount() int
	GetTgtSuffix() []string
}

type Handler interface {
	Init(HandlerConfigLoader) error
	CheckTarget(string) bool
	Action(string) error
}
