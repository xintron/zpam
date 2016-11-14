package zpam

// Handler manages Message event's.
//
// Handlers can be bound to commands or receive all type of events. Handlers
// will run in goroutines thereby allowing the handlers to do heavy work if
// need be.
type Handler interface {
	Handle(*Client, *Message)
}

// HandlerFunc adds a convenience wrapper around the Handler interface
type HandlerFunc func(*Client, *Message)

// Handle calls h(c, msg)
func (h HandlerFunc) Handle(c *Client, msg *Message) {
	h(c, msg)
}

// Make sure HandlerFunc implements the Handler interface
var _ Handler = (HandlerFunc)(nil)
