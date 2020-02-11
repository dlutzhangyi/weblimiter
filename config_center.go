package weblimiter

type ConfigCenter interface {
	GetConfig(key string) (string,error)
}
