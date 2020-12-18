package ziggurat

type MockKStreams struct {
	StartFunc func(z *Ziggurat) (chan struct{}, error)
}

func NewKafkaStreams() *MockKStreams {
	return &MockKStreams{StartFunc: func(z *Ziggurat) (chan struct{}, error) {
		return make(chan struct{}), nil
	}}
}

func (k MockKStreams) Start(z *Ziggurat) (chan struct{}, error) {
	return k.StartFunc(z)
}

func (k MockKStreams) Stop() {

}
