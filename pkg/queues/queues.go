package queues

import "servicer/pkg/cart"

// I'm not a fan of the name. It's a work in progress
// The requirements around a queue are relatively lax for now. This may not actually be populated in some cases, but instead acts as a stand in
// EG, direct kafka connections may just be their inputters and recipients, without its own queue struct
type QueueInt interface {
	NewInput() QueueInput
	NewRecipient() QueueRecipient
}

type QueueInput interface {
	// register this input with the queue
	Register() error
	// check with the queue if it is accepting inputs
	CanProceed() (bool, error)
	// submit an input to the queue
	Submit(*cart.Cart) error
	// close this connection with the queue
	Close() error
}

// a queue recipient,
type QueueRecipient interface {
	// register with the queue
	Register() error
	// returns the listener channel that will be called whenever the queue announces available carts
	CartCountListener() (<-chan (int), error)
	// request carts from the queue
	RequestData(int) ([]*cart.Cart, error)
	// close this connection to the queue
	Close() error
}
