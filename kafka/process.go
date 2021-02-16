package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/ziggurat"
)

func kafkaProcessor(msg *kafka.Message, route string, c *kafka.Consumer, h ziggurat.Handler, l ziggurat.StructuredLogger) {
	headers := map[string]string{ziggurat.HeaderMessageRoute: route, ziggurat.HeaderMessageType: "kafka"}
	event := CreateMessageEvent(msg.Value, headers)
	h.HandleEvent(event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err, nil)
}
