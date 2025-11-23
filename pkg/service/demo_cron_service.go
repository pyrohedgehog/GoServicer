package service

import (
	"fmt"
	"servicer/pkg/cart"
	"servicer/pkg/queues"
	"time"
)

type CronServiceConfig struct {
	CallFrequency    time.Duration     //TODO: this is the fastest way for now, but if this sticks around, this should be a better cron regulator
	OutputAttributes map[string]string //can be null, attached to the cart outputted
	OutputData       []byte            //the main data field of the output attribute
}
type CronService struct {
	config        ServiceConfig
	settings      CronServiceConfig
	ticker        *time.Ticker
	tickerRunning bool
	errChan       chan (error)
}

func (cs *CronService) Init(config ServiceConfig) error {
	// this can ignore any inputs configured, we only care about our outputs
	// TODO: we may want to throw an error on inputs being setup...

	// only accepting one output.
	if len(config.Outputs) != 1 {
		return fmt.Errorf("DemoCron only can output once. given %v", len(config.Outputs))
	}
	// setup the outputs and make sure they're registered
	for _, o := range config.Outputs {
		err := o.Register()
		if err != nil {
			return err
		}
	}
	if cs.errChan == nil {
		// TODO: this should become configurable, i think.
		cs.errChan = make(chan error, 128)
	}
	customData, ok := config.CustomData.(CronServiceConfig)
	if !ok {
		return fmt.Errorf("expected cron service config in custom data, got %T(%v)", config.CustomData, config.CustomData)
	}
	cs.settings = customData
	cs.config = config
	return nil
}
func (cs *CronService) Start() error {
	cs.ticker = time.NewTicker(cs.settings.CallFrequency)
	if !cs.tickerRunning {
		cs.tickerRunning = true

		// starts its asynchronous logic
		go func() {
			// stop really pauses the outputting to this, so no need to worry.
			for range cs.ticker.C {
				cs.OnTrigger(0)
			}
		}()
	}
	return nil
}

func (cs *CronService) Stop() {
	if cs.ticker != nil {
		cs.ticker.Stop()
	}
	// that really is all we needed to do.
}
func (cs *CronService) Close() error {
	cs.Stop()
	return nil
}
func (cs *CronService) OnTrigger(_ int) {
	// select the correct output, since we should only accept one
	var output queues.QueueInput
	for _, o := range cs.config.Outputs {
		if o == nil {
			continue
		}
		output = o
	}
	if output == nil {
		cs.errChan <- fmt.Errorf("could not run. No outputs")
		return
	}
	canProceed, err := output.CanProceed()
	if err != nil {
		// TODO: add more error data around this
		cs.errChan <- err
		return
	}
	if !canProceed {
		cs.errChan <- fmt.Errorf("could not run. No accepting outputs")
		return
	}

	// send out the event
	c := cart.NewCart()
	c.MainData = cs.settings.OutputData
	if cs.settings.OutputAttributes != nil {
		c.Attributes = cs.settings.OutputAttributes
	}
	err = output.Submit(c)
	if err != nil {
		cs.errChan <- err
	}
}
func (cs *CronService) GetErrorChan() <-chan (error) {
	return cs.errChan
}
