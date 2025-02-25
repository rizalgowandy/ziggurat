package kafka

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/gojekfarm/ziggurat"
)

// try with an existing topic

func makeRandString() string {
	bb := make([]byte, 5)
	_, err := rand.Read(bb)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", bb)
}

func Test_streams(t *testing.T) {
	c, cfn := context.WithCancel(context.Background())
	time.AfterFunc(time.Millisecond*3000, func() { cfn() })
	defer cfn()
	var expectedMessageCount int32 = 5
	topic := makeRandString()
	var messageCount int32
	done := make(chan struct{})
	ks := Streams{
		StreamConfig: StreamConfig{{
			BootstrapServers: "localhost:9092",
			GroupID:          topic + "_consumer",
			ConsumerCount:    1,
			Topics:           topic,
			AutoOffsetReset:  "earliest",
		}},
	}
	go func() {
		err := ks.Stream(c, ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) error {
			atomic.AddInt32(&messageCount, 1)
			return nil
		}))

		if !errors.Is(err, ErrCleanShutdown) {
			t.Errorf("streams failed with error:%v", err)
		}
		done <- struct{}{}
	}()

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		t.Errorf("could not create producer:%v", err)
	}
	deliveryCh := make(chan kafka.Event)
	go func() {
		for e := range deliveryCh {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					t.Errorf("delivery failed with error:%v", err)
				} else {
					t.Logf("delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	for i := 0; i < 5; i++ {
		err := p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
			},
			Value:     []byte("foo"),
			Key:       []byte("foo"),
			Timestamp: time.Now(),
		}, deliveryCh)
		if err != nil {
			t.Errorf("error producing:%v", err)
		}
	}

	<-done
	p.Close()
	got := atomic.LoadInt32(&messageCount)
	if expectedMessageCount != got {
		t.Errorf("expected message count [%d] got [%d]", expectedMessageCount, got)
	}
}
