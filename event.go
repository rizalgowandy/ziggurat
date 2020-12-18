package ziggurat

import (
	"sync"
	"time"
)

type MessageEvent struct {
	MessageValue    []byte
	MessageKey      []byte
	Topic           string
	StreamRoute     string
	ActualTimestamp time.Time
	TimestampType   string
	Attributes      map[string]interface{}
	attrMutex       *sync.Mutex
	//exposes Attributes for gob encoding, use Get and Set for thread safety
}

func NewMessageEvent(key []byte, value []byte, topic string, route string, timestampType string, ktimestamp time.Time) MessageEvent {
	return MessageEvent{
		Attributes:      map[string]interface{}{},
		attrMutex:       &sync.Mutex{},
		MessageValue:    value,
		MessageKey:      key,
		Topic:           topic,
		StreamRoute:     route,
		TimestampType:   timestampType,
		ActualTimestamp: ktimestamp,
	}
}

func (m MessageEvent) GetMessageAttribute(key string) interface{} {
	m.attrMutex.Lock()
	defer m.attrMutex.Unlock()
	return m.Attributes[key]
}

func (m *MessageEvent) SetMessageAttribute(key string, value interface{}) {
	m.attrMutex.Lock()
	defer m.attrMutex.Unlock()
	m.Attributes[key] = value
}
