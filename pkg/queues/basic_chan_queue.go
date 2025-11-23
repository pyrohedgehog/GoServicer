package queues

import (
	"fmt"
	"servicer/pkg/cart"

	"github.com/google/uuid"
)

// this is the most simple queue possible. Primarily built to be a proof of concept item.
type BasicChanQueue struct {
	inputBufferSize  int                      //the buffer size for the input channel. can only be set when there are no input connections
	outputBufferSize int                      //the default buffer size for output buffers
	inputCount       int                      //how many inputs to this are open
	outputCount      int                      //how many outputs from this are open
	input            chan (*cart.Cart)        //the single input point for all data
	outputCalls      []chan (int)             //an array of all the output points channels
	queuedData       map[uuid.UUID]*cart.Cart //the data related to each id
	queue            []uuid.UUID              //the order of the ids
}

func NewBasicChanQueue() *BasicChanQueue {
	bcq := &BasicChanQueue{
		inputBufferSize:  512,
		outputBufferSize: 256,
		outputCalls:      []chan int{},
		queuedData:       map[uuid.UUID]*cart.Cart{},
	}
	return bcq
}

func (bcq *BasicChanQueue) Start() {
	go func() {
		for i := range bcq.input {
			bcq.queuedData[i.Id] = i
			bcq.queue = append(bcq.queue, i.Id)
			// TODO: select target logic should be added. for now, always the first available
			for _, outChan := range bcq.outputCalls {
				if outChan != nil {
					outChan <- len(bcq.queuedData)
					break
				}
			}
		}
		fmt.Println("queue closing, should be restarted")
	}()
}
func (bcq *BasicChanQueue) pop(n int) ([]*cart.Cart, error) {
	if n > len(bcq.queuedData) {
		// TODO: add useful error data
		return nil, fmt.Errorf("err data to be put here")
	}
	var ansIds []uuid.UUID
	ansIds, bcq.queue = bcq.queue[:n], bcq.queue[n:]
	// get the data from the queues
	ans := make([]*cart.Cart, 0, len(ansIds))
	for _, id := range ansIds {
		ans = append(ans, bcq.queuedData[id])
		delete(bcq.queuedData, id)
	}
	return ans, nil
}
func (bcq *BasicChanQueue) NewInput() QueueInput {
	return &basicChanQueueInput{
		src:    bcq,
		stream: make(chan *cart.Cart),
	}
}
func (bcq *BasicChanQueue) NewRecipient() QueueRecipient {
	return &basicChanQueueRecipient{
		src: bcq,
	}
}
func (bcq *BasicChanQueue) AddInput() chan (*cart.Cart) {
	if bcq.input == nil {
		bcq.input = make(chan *cart.Cart, bcq.inputBufferSize)
		bcq.Start()
	}
	bcq.inputCount++
	return bcq.input
}
func (bcq *BasicChanQueue) RemoveInput() {
	bcq.inputCount--
	if bcq.inputCount == 0 {
		close(bcq.input)
	}
}
func (bcq *BasicChanQueue) AddOutput() (id int, stream <-chan (int)) {
	newChan := make(chan (int), bcq.outputBufferSize)
	bcq.outputCalls = append(bcq.outputCalls, newChan)
	bcq.outputCount++
	return len(bcq.outputCalls) - 1, newChan
}
func (bcq *BasicChanQueue) RemoveOutput(id int) error {
	// TODO: check this id!
	bcq.outputCount--
	close(bcq.outputCalls[id])
	bcq.outputCalls[id] = nil
	return nil
}

type basicChanQueueInput struct {
	src    *BasicChanQueue
	stream chan (*cart.Cart)
}

func (qi *basicChanQueueInput) Register() error {
	// prevent duplicate registering
	if cap(qi.stream) == 0 {
		qi.stream = qi.src.AddInput()
	}
	return nil
}

// the basic channel queue input assumes it will always be able to ingest data
func (qi *basicChanQueueInput) CanProceed() (bool, error) {
	return true, nil
}
func (qi *basicChanQueueInput) Submit(c *cart.Cart) error {
	qi.stream <- c
	return nil
}
func (qi *basicChanQueueInput) Close() error {
	qi.src.RemoveInput()
	qi.stream = nil
	return nil
}

type basicChanQueueRecipient struct {
	src    *BasicChanQueue
	stream <-chan (int)
	id     int
}

func (qr *basicChanQueueRecipient) Register() error {
	qr.id, qr.stream = qr.src.AddOutput()
	return nil
}
func (qr *basicChanQueueRecipient) CartCountListener() (<-chan (int), error) {
	return qr.stream, nil
}
func (qr *basicChanQueueRecipient) RequestData(n int) ([]*cart.Cart, error) {
	return qr.src.pop(n)
}
func (qr *basicChanQueueRecipient) Close() error {
	err := qr.src.RemoveOutput(qr.id)
	if err != nil {
		return err
	}
	qr.stream = nil
	return nil
}
