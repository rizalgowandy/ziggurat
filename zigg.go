package ziggurat

import (
	"context"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/sethvargo/go-signalcontext"
	"sync/atomic"
	"syscall"
)

type Ziggurat struct {
	handler   Handler
	Logger    StructuredLogger
	startFunc StartFunction
	stopFunc  StopFunction
	isRunning int32
	streams   Streamer
}

func (z *Ziggurat) Run(ctx context.Context, streams Streamer, handler Handler) chan error {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return nil
	}

	if z.Logger == nil {
		z.Logger = logger.NewJSONLogger("info")
	}

	if streams == nil {
		panic("`kafka` cannot be nil")
	}

	if handler == nil {
		panic("`handler` cannot be nil")
	}

	z.streams = streams

	doneChan := make(chan error)
	parentCtx, canceler := signalcontext.Wrap(ctx, syscall.SIGINT, syscall.SIGTERM)

	z.handler = handler

	atomic.StoreInt32(&z.isRunning, 1)
	go func() {
		err := <-z.start(parentCtx, z.startFunc)
		z.Logger.Error("error starting kafka", err)
		canceler()
		atomic.StoreInt32(&z.isRunning, 0)
		z.stop(z.stopFunc)
		doneChan <- err
	}()
	return doneChan
}

func (z *Ziggurat) start(ctx context.Context, startCallback StartFunction) chan error {
	if startCallback != nil {
		z.Logger.Info("invoking start function")
		startCallback(ctx)
	}

	streamsStop := z.streams.Stream(ctx, z.handler)
	return streamsStop
}

func (z *Ziggurat) stop(stopFunc StopFunction) {
	if stopFunc != nil {
		z.Logger.Info("invoking stop function")
		stopFunc()
	}
}

func (z *Ziggurat) IsRunning() bool {
	if atomic.LoadInt32(&z.isRunning) == 1 {
		return true
	}
	return false
}

func (z *Ziggurat) StartFunc(function StartFunction) {
	z.startFunc = function
}

func (z *Ziggurat) StopFunc(function StopFunction) {
	z.stopFunc = function
}
