package weblimiter

type RateLimiter interface {
	//decide to limit a request
	Limited(request string, options map[string]string) bool
}
