package cart

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	// this is the main data to be used.
	MainData []byte
	// this will be useful for knowing if a cart has changed its data, without doing full comparisons
	Checksum []byte
	// the ID of this data cart. Will likely be unchecked for the longest time until we want to log anything
	Id uuid.UUID
	// attributes associated with this cart. This I'm fine limiting to a key value pairing. Fastest to sort by, fasted to change.
	Attributes map[string]string
	CreatedAt  time.Time //when this was first created
	UpdatedAt  time.Time //when this was last updated
	// TODO: parents, lineage, all these other things
	// TODO: I may want to add a extended data field to represent larger data stored elsewhere, which is only passed on request (EG, file gets uploaded, you want to avoid moving that large data around, and would instead want to pass a reader of it)
}

func NewCart() *Cart {
	c := &Cart{
		MainData:   []byte{},
		Checksum:   []byte{},
		Id:         uuid.New(),
		Attributes: map[string]string{},
		CreatedAt:  time.Now().In(time.UTC),
	}
	return c
}

// get a string format. Ideal for printing
func (c Cart) ToString() string {
	return fmt.Sprintf("Id:%v, attributes:%v, main_data:\"%s\", checksum:%v", c.Id, c.Attributes, c.MainData, c.Checksum)
}
func (c *Cart) Close() error {
	// for now, this is useless, but if we add file references, or things like that, it'd be useful to close any trailing connections
	return nil
}
