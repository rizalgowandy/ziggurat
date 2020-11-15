package zig

import (
	"context"
	testproto "github.com/gojekfarm/ziggurat-go/protobuf"
	"github.com/golang/protobuf/proto"
	"reflect"
	"testing"
	"time"
)

type JSONMessage struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SecretNum int    `json:"secretNum"`
}

type mwMockApp struct {
	t *testing.T
}

func (m mwMockApp) ConfigReader() ConfigReader {
	panic("implement me")
}

func (m mwMockApp) Context() context.Context {
	return nil
}

func (m mwMockApp) Router() StreamRouter {
	return nil
}

func (m mwMockApp) MessageRetry() MessageRetry {
	return nil
}

func (m mwMockApp) Run(router StreamRouter, options RunOptions) chan struct{} {
	return nil
}

func (m mwMockApp) Configure(configFunc func(o App) Options) {

}

func (m mwMockApp) MetricPublisher() MetricPublisher {
	return &mockMetrics{m.t}
}

func (m mwMockApp) HTTPServer() HttpServer {
	return nil
}

func (m mwMockApp) Config() *Config {
	return nil
}

func (m mwMockApp) Stop() {

}

func (m mwMockApp) IsRunning() bool {
	return false
}

type mockMetrics struct {
	t *testing.T
}

func (m mockMetrics) Start(app App) error {
	return nil
}

func (m mockMetrics) Stop() error {
	return nil
}

func (m mockMetrics) IncCounter(metricName string, value int64, args map[string]string) error {
	expectedArgs := map[string]string{
		"topic_entity": "te-1",
		"kafka_topic":  "topic-1",
	}
	expectedMetricName := "message_count"
	expectedValue := int64(1)
	if value != expectedValue {
		m.t.Errorf("expected %d, got %d", expectedValue, value)
	}
	if !reflect.DeepEqual(expectedArgs, args) {
		m.t.Error("args mismatch")
	}
	if metricName != expectedMetricName {
		m.t.Errorf("expected metric name to be %s, but got %s", "message_count", metricName)
	}
	return nil
}

func (m mockMetrics) Gauge(metricName string, value int64, arguments map[string]string) error {
	expectedArgs := map[string]string{
		"topic_entity": "te-1",
		"kafka_topic":  "topic-1",
	}
	expectedValue := int64(0)
	expectedMetricName := "message_delay"
	if metricName != expectedMetricName {
		m.t.Errorf("expected metric name to be %s, but got %s", "message_delay", metricName)
	}
	if !reflect.DeepEqual(expectedArgs, arguments) {
		m.t.Errorf("expected metric name %s got %s", expectedMetricName, metricName)
	}

	if expectedValue != value {
		m.t.Errorf("expected value %d, got %d", expectedValue, value)
	}
	return nil
}

func TestJSONDeserializer(t *testing.T) {
	jsonMessage := `{"id":"xyzzyspoonshift1","name":"road_rash","secretNum":1}`
	expectedMessage := JSONMessage{
		ID:        "xyzzyspoonshift1",
		Name:      "road_rash",
		SecretNum: 1,
	}
	handler := func(event MessageEvent, app App) ProcessStatus {
		msg := &JSONMessage{}
		event.DecodeValue(msg)
		if *msg == expectedMessage {
			return ProcessingSuccess
		}
		return SkipMessage
	}
	result := JSONDecoder(handler)(MessageEvent{
		MessageValueBytes: []byte(jsonMessage),
	}, &mwMockApp{})

	if result == SkipMessage {
		t.Errorf("expected %v but got %v", ProcessingSuccess, result)
	}
}

func TestProtobufDeserializer(t *testing.T) {
	expectedMessage := testproto.TestMessage{
		Id:        "1",
		SecretNum: 0,
		Name:      "1",
	}
	bytes, _ := proto.Marshal(&expectedMessage)
	//go vet complains copying lockers by value
	ProtoDecoder(func(messageEvent MessageEvent, app App) ProcessStatus {
		testProtoModel := testproto.TestMessage{}
		messageEvent.DecodeValue(&testProtoModel)
		if !proto.Equal(&testProtoModel, &expectedMessage) {
			t.Errorf("proto messages are not equal")
		}
		return ProcessingSuccess
	})(MessageEvent{
		MessageValueBytes: bytes,
	}, &mwMockApp{})

}

func TestMessageLogger(t *testing.T) {
	expectedMsg := "Msg logger middleware"
	ts := time.Time{}
	messageVal := []byte("foo bar")
	topicEntity := "te-1"
	kafkaTopic := "topic-1"
	expectedArgs := map[string]interface{}{
		"topic-entity":  topicEntity,
		"kafka-topic":   kafkaTopic,
		"kafka-ts":      ts.String(),
		"message-value": string(messageVal),
	}
	origLogInfo := logInfo
	defer func() {
		logInfo = origLogInfo
	}()
	logInfo = func(msg string, args map[string]interface{}) {
		if msg != expectedMsg {
			t.Errorf("expected msg to be %s but got %s", expectedMsg, msg)
		}
		if !reflect.DeepEqual(expectedArgs, args) {
			t.Errorf("expected args to be %+v but got %+v", expectedArgs, args)
		}
	}
	msgLogger := MessageLogger(func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	})

	msgLogger(MessageEvent{
		MessageValueBytes: messageVal,
		MessageKeyBytes:   nil,
		Topic:             kafkaTopic,
		TopicEntity:       topicEntity,
		KafkaTimestamp:    time.Time{},
		TimestampType:     "",
		Attributes:        nil,
	}, &mwMockApp{})
}

func TestMessageMetricsPublisher(t *testing.T) {
	origGetCurrTime := getCurrTime
	getCurrTime = func() time.Time {
		return time.Time{}
	}
	defer func() {
		getCurrTime = origGetCurrTime
	}()
	messageMetricsPublisher := MessageMetricsPublisher(func(messageEvent MessageEvent, app App) ProcessStatus {
		return ProcessingSuccess
	})
	messageMetricsPublisher(MessageEvent{
		MessageValueBytes: nil,
		MessageKeyBytes:   nil,
		Topic:             "topic-1",
		TopicEntity:       "te-1",
		KafkaTimestamp:    time.Time{},
		TimestampType:     "",
		Attributes:        nil,
	}, &mwMockApp{t})
}
