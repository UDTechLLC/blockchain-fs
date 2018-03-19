package core

type StorageApi interface {
	Create(origin string) (exitCode int, err error)
	Delete(origin string) (exitCode int, err error)
	Mount(origin string, notifypid int) (exitCode int, err error)
	Unmount(origin string) (exitCode int, err error)
}
