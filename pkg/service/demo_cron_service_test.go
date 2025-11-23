package service

import (
	"servicer/pkg/queues"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOnTrigger(t *testing.T) {
	queue := queues.NewBasicChanQueue()
	// var queue = queues.NewBasicChanQueue()
	input := queue.NewInput()
	assert.NoError(
		t,
		input.Register(),
	)
	config := ServiceConfig{
		Name:    "test service",
		Outputs: map[string]queues.QueueInput{"success": input},
		CustomData: CronServiceConfig{
			CallFrequency:    time.Second,
			OutputAttributes: map[string]string{},
			OutputData:       []byte("Hello World!"),
		},
	}
	cs := &CronService{}
	assert.NoError(
		t,
		cs.Init(config),
	)
	errChan := cs.GetErrorChan()
	assert.Len(t, errChan, 0)
	cs.OnTrigger(0)
	assert.Len(t, errChan, 0)
}
