package weblimiter

type ConfigClient interface {
	GetConfig(key string) (map[string]string, error)
	ParseConfig(config map[string]string) ([]RateConf, error)
	RegisterConfigChannel(ch chan []RateConf)
	Daemon()
}
