package weblimiter

type MockClient struct {
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) GetConfig(key string) (map[string]string, error) {
	return nil, nil
}
func (c *MockClient) ParseConfig(config map[string]string) ([]RateConf, error) {
	return nil, nil
}
func (c *MockClient) RegisterConfigChannel(ch chan []RateConf) {

}
func (c *MockClient) Daemon() {

}
