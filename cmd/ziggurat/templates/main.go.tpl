package main

import (
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw/statsd"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	statsdPub := statsd.NewPublisher(statsd.WithDefaultTags(map[string]string{
		"app_name": "{{.AppName}}",
	}))
	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	ks := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				Topics:           "plain-text-log",
				GroupID:          "{{.AppName}}_consumer",
				ConsumerCount:    1,
				RouteGroup:       "plain-text-group",
			},
		},
		Logger: l,
	}

	r.HandleFunc("plain-text-group/*", func(ctx context.Context, event *ziggurat.Event) error {
		return nil
	})


	zig.StartFunc(func(ctx context.Context) {
		err := statsdPub.Run(ctx)
		l.Error("statsd publisher error", err)
	})

	if runErr := zig.RunAll(ctx, &r, &ks); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
