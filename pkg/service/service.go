package service

import "servicer/pkg/queues"

// The service is the bit actually handling each cart of data.
type Service interface {
	Init(ServiceConfig) error   //called once, startup the service for the first time
	Start() error               //time to start up this service. Possibly restarting from a stopped position
	Stop()                      //time to stop this service, May still be restarted, but should not have normal async (clock like) functions running
	Close() error               //time to shutdown this application. May not be restarted.
	OnTrigger(count int)        //called when new inputs are available, or is told to run once
	GetErrorChan() chan (error) //the error channel, where all errors are output for this service
}
type ServiceConfig struct {
	Name    string
	Input   queues.QueueRecipient
	Outputs map[string]queues.QueueInput //a name given for each output in the queue

	CustomData any //any custom data required by your service can be configured here if not handled otherwise
}
