package config

type DelLog interface {
	GetDelLogTask() []string
	GetDelLogLogDir(task string) (bool, string)
	GetDelLogLogExpiredTime(task string) (bool, int)
	GetDelLogInterval(task string) (bool, int)
}
