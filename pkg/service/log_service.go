package service

import (
	"fmt"
	"time"
)

// a simple log service. It receives as many carts as given, and logs them (prints to the console), then outputs a copy to all downstream queues
type LogService struct {
	config   ServiceConfig
	settings LogServiceConfig
	errChan  chan (error)
}

type LogServiceConfig struct {
	BatchReadSize int //how many carts are batched in a read. Default 1
}

func (ls *LogService) Init(config ServiceConfig) error {
	// parse the config
	customData, ok := config.CustomData.(LogServiceConfig)
	if !ok {
		return fmt.Errorf("expected log service config in custom data, got %T(%v)", config.CustomData, config.CustomData)
	}
	ls.settings = customData
	ls.config = config
	// set the default
	if ls.settings.BatchReadSize == 0 {
		ls.settings.BatchReadSize = 1
	}
	// setup the outputs and make sure they're registered
	for _, o := range config.Outputs {
		err := o.Register()
		if err != nil {
			return err
		}
	}

	if ls.errChan == nil {
		// TODO: this should become configurable, i think.
		ls.errChan = make(chan error, 128)
	}
	return nil
}
func (ls *LogService) Start() error {
	// everything's setup by the listener
	err := ls.config.Input.Register()
	if err != nil {
		return err
	}
	listener, err := ls.config.Input.CartCountListener()
	if err != nil {
		return err
	}
	go func(listener <-chan (int)) {
		for val := range listener {
			ls.OnTrigger(val)
		}
	}(listener)
	return nil
}

func (ls *LogService) Stop() {
	// not much we need to do but tell the input we're done.
	err := ls.config.Input.Close()
	if err != nil {
		ls.errChan <- err
	}
}
func (ls *LogService) Close() error {
	ls.Stop()
	return nil
}
func (ls *LogService) OnTrigger(count int) {
	// see if we should get the carts or if the batch is too small
	if count < ls.settings.BatchReadSize {
		// not enough in queue
		return
	}
	// get cart data
	inputCarts, err := ls.config.Input.RequestData(ls.settings.BatchReadSize)
	if err != nil {
		ls.errChan <- err
		return
	}
	// do the actual logging
	for _, c := range inputCarts {
		fmt.Printf("%v-Logging@%v: %v\n", ls.config.Name, time.Now().Format(time.DateTime), c.ToString())
	}

	// send a copy to all outputs
	for name, o := range ls.config.Outputs {
		if o == nil {
			continue
		}
		for _, cart := range inputCarts {
			canProceed, err := o.CanProceed()
			if err != nil {
				// TODO: add more error data around this
				ls.errChan <- err
				continue
			}
			if !canProceed {
				ls.errChan <- fmt.Errorf("output %v not accepting data", name)
				continue
			}

			err = o.Submit(cart)
			if err != nil {
				// TODO: add more error data around this
				ls.errChan <- err
			}
		}

	}
}
func (ls *LogService) GetErrorChan() <-chan (error) {
	return ls.errChan
}
