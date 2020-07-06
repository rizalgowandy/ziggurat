package ziggurat

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"strings"
	"sync"
)

type topicEntity struct {
	handlerFunc      HandlerFunc
	consumers        []*kafka.Consumer
	bootstrapServers string
	originTopics     []string
}

type TopicEntityHandlerMap = map[string]topicEntity

type InstanceCount = map[string]int

type StreamRouter struct {
	handlerFunctionMap map[TopicEntityName]*topicEntity
}

func newConsumerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	}
}

func NewStreamRouter() *StreamRouter {
	return &StreamRouter{
		handlerFunctionMap: make(map[TopicEntityName]*topicEntity),
	}
}

func (sr *StreamRouter) HandlerFunc(topicEntityName TopicEntityName, handlerFn HandlerFunc) {
	sr.handlerFunctionMap[topicEntityName] = &topicEntity{handlerFunc: handlerFn}
}

func makeKV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func (sr *StreamRouter) Start(srConfig StreamRouterConfigMap) {
	hfMap := sr.handlerFunctionMap
	if len(hfMap) == 0 {
		log.Fatal("unable to start stream-router, no handler functions registered")
	}
	var wg sync.WaitGroup
	for topicEntityName, topicEntity := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]
		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", strings.Join(streamRouterCfg.BootstrapServers, ","))
		groupID := makeKV("group.id", streamRouterCfg.GroupID)

		consumerConfig.Set(bootstrapServers)
		consumerConfig.Set(groupID)
		consumers := StartConsumers(consumerConfig, topicEntityName, streamRouterCfg.OriginTopics, streamRouterCfg.InstanceCount, topicEntity.handlerFunc, &wg)
		
		topicEntity.originTopics = streamRouterCfg.OriginTopics
		topicEntity.consumers = consumers
	}
	wg.Wait()
}
