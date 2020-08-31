### Ziggurat GO [WIP]

### Installation
- Configure git for using private modules
```shell script
git config --global url."git@source.golabs.io:".insteadOf "https://source.golabs.io/"
```
- Run go get
```shell script
go get -v -u source.golabs.io/lambda/zigg-go/ziggurat                                                                                                                                                          
```

#### How to use
- create a `config/config.yaml` in your project root
- sample `config.yaml`
```yaml
service-name: "test-service"
stream-router:
  test-entity2:
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "test-topic1"
    group-id: "test-group"
    topic-entity: "topic-entiy2"
  test-entity:
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "test-topic2"
    group-id: "test-group2"
    topic-entity: "test-entity"
  json-entity:
    bootstrap-servers: "localhost:9092"
    instance-count: 1
    origin-topics: "json-test"
    group-id: "json-group"
    topic-entity: "json-entity"
log-level: "debug"
retry:
  enabled: true
  count: 5
rabbitmq:
  host: "amqp://user:bitnami@localhost:5672/"
  delay-queue-expiration: "1000"
```

- sample `main.go`

```go
package main

import (
	"fmt"
	"source.golabs.io/lambda/zigg-go/ziggurat"
)

func main() {
	router := ziggurat.NewStreamRouter()
	router.HandlerFunc("booking", func(messageEvent ziggurat.MessageEvent) ziggurat.ProcessStatus {
		fmt.Println("Message -> ", messageEvent)
		return ziggurat.ProcessingSuccess
	})

	ziggurat.Start(router, ziggurat.StartupOptions{
		StartFunction: func(config ziggurat.Config) {
			fmt.Println("Start function called...")
		},
		StopFunction: func() {
			fmt.Println("Stopping app...")
		},
		Retrier: nil,
	})

}
```
 


#### TODO
- [x] Balanced Consumer groups
- [x] RabbitMQ retries
- [x] Atleast once delivery semantics
- [x] Retry interface
- [x] Default middleware to deserialize messages
- [x] Env vars Config override
- [ ] Replay RabbitMQ deadset messages
- [ ] Log formatting
- [ ] Configurable RabbitMQ consumer count
- [ ] Unit tests
