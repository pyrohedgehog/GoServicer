package queues

import (
	"servicer/pkg/cart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasics(t *testing.T) {
	bcq := NewBasicChanQueue()
	in := bcq.NewInput()
	assert.NoError(
		t,
		in.Register(),
	)
	out := bcq.NewRecipient()
	assert.NoError(
		t,
		out.Register(),
	)

	outCaller, err := out.CartCountListener()
	assert.NoError(t, err)
	assert.Len(t, outCaller, 0)

	// setup some test data
	testData := cart.NewCart()
	testData.Attributes["foo"] = "bar"
	// see if we can pass the data
	canProceed, err := in.CanProceed()
	assert.NoError(t, err)
	assert.True(t, canProceed)
	// pass the data
	err = in.Submit(testData)
	assert.NoError(t, err)

	// wait till we get something.
	outCount := <-outCaller
	assert.Equal(t, 1, outCount)

	data, err := out.RequestData(1)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, testData, data[0])

}
