package main

import (
	"fmt"
	"servicer/pkg/queues"
	"servicer/pkg/service"
	"time"
)

func main() {
	cronToLog := queues.NewBasicChanQueue()
	qIn := cronToLog.NewInput()
	qOut := cronToLog.NewRecipient()
	cron := &service.CronService{}
	err := cron.Init(service.ServiceConfig{
		Name:    "test service",
		Outputs: map[string]queues.QueueInput{"success": qIn},
		CustomData: service.CronServiceConfig{
			CallFrequency:    time.Second * 3,
			OutputAttributes: map[string]string{},
			OutputData:       []byte("Hello World!"),
		},
	})
	if err != nil {
		panic(err)
	}
	logger := &service.LogService{}
	err = logger.Init(service.ServiceConfig{
		Name:    "LogService",
		Input:   qOut,
		Outputs: map[string]queues.QueueInput{},
		CustomData: service.LogServiceConfig{
			BatchReadSize: 1,
		},
	})
	if err != nil {
		panic(err)
	}
	cronErrs := cron.GetErrorChan()
	logErrs := cron.GetErrorChan()
	go func() {
		// just always print any errors
		for {
			select {
			case err := <-cronErrs:
				fmt.Printf("cronErr: %v", err)
			case err := <-logErrs:
				fmt.Printf("logErr: %v", err)
			}
		}
	}()
	err = logger.Start()
	if err != nil {
		panic(err)
	}
	cron.OnTrigger(1)
	// get the cron to run once
	// assume logger printed it
	<-time.After(time.Second)
	fmt.Println("log should have printed by now")
	// start the auto run
	err = cron.Start()
	defer cron.Close()
	defer logger.Close()
	if err != nil {
		panic(err)
	}
	// running, assuming we'll see some items logged
	<-time.After(time.Second * 16)
	fmt.Println("log should've logged more, right?")

}
