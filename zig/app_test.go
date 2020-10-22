package zig

import (
	"github.com/julienschmidt/httprouter"
	"testing"
)

var app *App
var mhttp, mrouter, mstatsd, mrabbitmq = &mockHTTP{}, &mockRouter{}, &mockStatsD{}, &mockRabbitMQ{}
var startCount = 0
var stopCount = 0
var expectedStopCount = 3
var expectedStartCount = 4

type mockHTTP struct{}
type mockStatsD struct{}
type mockRouter struct{}
type mockRabbitMQ struct{}

func (m *mockRabbitMQ) Start(app *App) (chan int, error) {
	startCount++
	stopChan := make(chan int)
	go func() {
		close(stopChan)
	}()
	return stopChan, nil
}

func (m *mockRabbitMQ) Retry(app *App, payload MessageEvent) error {
	return nil
}

func (m *mockRabbitMQ) Stop() error {
	stopCount++
	return nil
}

func (m *mockRabbitMQ) Replay(app *App, topicEntity string, count int) error {
	return nil
}

func (m *mockStatsD) Start(app *App) error {
	startCount++
	return nil
}

func (m *mockStatsD) Stop() error {
	stopCount++
	return nil
}

func (m *mockStatsD) Gauge(metricName string, value int64, arguments map[string]string) error {
	return nil
}

func (m *mockStatsD) IncCounter(metricName string, value int64, arguments map[string]string) error {
	return nil
}

func (m *mockRouter) Start(app *App) (chan int, error) {
	startCount++
	closeChan := make(chan int)
	go func() {
		close(closeChan)
	}()
	return closeChan, nil
}

func (m *mockRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc, mw ...MiddlewareFunc) {

}

func (m *mockRouter) GetTopicEntities() []*topicEntity {
	return []*topicEntity{}
}

func (m *mockRouter) GetHandlerFunctionMap() map[string]*topicEntity {
	return map[string]*topicEntity{}
}

func (mh *mockHTTP) Start(app *App) {
	startCount++
}

func (mh *mockHTTP) attachRoute(func(r *httprouter.Router)) {

}

func (mh *mockHTTP) Stop() error {
	stopCount++
	return nil
}

func setup() {
	app = &App{}
	app.router = mrouter
	app.httpServer = mhttp
	app.metricPublisher = mstatsd
	app.retrier = mrabbitmq
	app.cancelFun = func() {}
}

func teardown() {
	app = &App{}
	startCount = 0
	stopCount = 0
}

func TestApp_Start(t *testing.T) {
	setup()
	defer teardown()
	startCallbackCalled := false
	startCallback := func(app *App) {
		startCallbackCalled = true
	}

	app.start(startCallback, nil)

	if startCount < expectedStartCount {
		t.Errorf("expected start count to be %v but got %v", expectedStartCount, startCount)
	}

	if !startCallbackCalled {
		t.Errorf("expected startCallbackCalled to be %v, but got %v", true, startCallbackCalled)
	}

}

func TestApp_Stop(t *testing.T) {
	setup()
	defer teardown()
	stopCallbackCalled := false

	app.stop(func() {
		stopCallbackCalled = true
	})
	if stopCount < expectedStopCount {
		t.Errorf("expected stop count to be %v, but got %v", expectedStopCount, stopCount)
	}
	if !stopCallbackCalled {
		t.Errorf("expected stopCallbackCalled to be %v, but got %v", true, stopCallbackCalled)
	}
}
