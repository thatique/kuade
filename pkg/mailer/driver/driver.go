package driver

import (
	"context"

	"github.com/emersion/go-message"
	"github.com/thatique/kuade/pkg/kerr"
)

type Transport interface {
	//  Open a network connection.
	Open(context.Context) error

	// Close a network connection.
	Close(context.Context) error

	// Send one or more message.Entity objects and return the number of email
	// messages sent. Before calling this method, you need to make sure to call
	// Open, otherwise most implementation might choose to always return 0.
	SendMessages(context.Context, []*message.Entity) (int, error)

	// ErrorCode should return a code that describes the error, which was returned by
	// one of the other methods in this interface.
	ErrorCode(err error) kerr.ErrorCode
}
