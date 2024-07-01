package filestream_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	. "github.com/wandb/wandb/core/internal/filestream"
	"github.com/wandb/wandb/core/internal/waitingtest"
)

func TestTransmitLoop_Sends(t *testing.T) {
	outputs := make(chan *FileStreamRequestJSON)
	loop := TransmitLoop{
		HeartbeatStopwatch:     waitingtest.NewFakeStopwatch(),
		LogFatalAndStopWorking: func(err error) {},
		Send: func(
			ftd *FileStreamRequestJSON,
			c chan<- map[string]any,
		) error {
			outputs <- ftd
			return nil
		},
	}
	testInput := NewRequestReader(&FileStreamRequest{Preempting: true})

	inputs := make(chan *FileStreamRequestReader)
	_ = loop.Start(inputs, FileStreamOffsetMap{})
	inputs <- testInput
	close(inputs)

	select {
	case result := <-outputs:
		assert.True(t, *result.Preempting)
	case <-time.After(time.Second):
		t.Error("timeout after 1 second")
	}
}

func TestTransmitLoop_SendsHeartbeats(t *testing.T) {
	heartbeat := waitingtest.NewFakeStopwatch()
	inputs := make(chan *FileStreamRequestReader)
	defer close(inputs)
	outputs := make(chan *FileStreamRequestJSON)
	loop := TransmitLoop{
		HeartbeatStopwatch:     heartbeat,
		LogFatalAndStopWorking: func(err error) {},
		Send: func(
			ftd *FileStreamRequestJSON,
			c chan<- map[string]any,
		) error {
			outputs <- ftd
			return nil
		},
	}

	loop.Start(inputs, FileStreamOffsetMap{})
	heartbeat.SetDone()

	select {
	case result := <-outputs:
		assert.Zero(t, *result)
	case <-time.After(time.Second):
		t.Error("timeout after 1 second")
	}
}
