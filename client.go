package zpam

import (
	"errors"
	"log"
	"strings"
)

var (
	// ErrExistingCommand returned if the same command is added more than once
	ErrExistingCommand = errors.New("zpam: command already exists")
	// ErrNilHandler returned when the handler is nil
	ErrNilHandler = errors.New("zpam: nil handler")
	// ErrEmptyCommand returned when an empty command is given
	ErrEmptyCommand = errors.New("zpam: empty command")
	// ErrUnavailableBackend returned when the backend requested isn't available
	ErrUnavailableBackend = errors.New("zpam: unavailable backend")
)

// Client takes care of commands management and backend connections.
type Client struct {
	Name   string
	Prefix string

	// Commands will be routed within the Client and only dispatched to
	// handlers that have the given Client.Prefix + command-string setup
	commands map[string]Handler
	// handlers that will receive all events
	handlers []Handler
	// callbacks when server is stopping
	onShutdown []func() error
	// Active backend
	backend Backend
}

func (c *Client) Start(backend string) error {
	init, ok := backends[backend]
	if !ok {
		return ErrUnavailableBackend
	}
	c.backend = init(c)
	return nil
}

func (c *Client) Send(to, text string) error {
	c.backend.Send(to, &Message{Text: text})
	return nil
}

// Receive accepts incoming Message's from backends.
//
// A message needs to be constructed so that it can be used by the different
// available backends.
func (c *Client) Receive(msg *Message) error {
	var h Handler
	// Dispatch to all global handlers
	for _, h = range c.handlers {
		h.Handle(c, msg)
	}

	cmd := c.parseCommand(msg)
	h = c.commands[cmd]

	// command handler registered, run the handler
	if h != nil {
		h.Handle(c, msg)
	}
	return nil
}

func (c *Client) parseCommand(msg *Message) string {
	// Begins with the prefix character
	if strings.HasPrefix(msg.Text, c.Prefix) {
		cmd := strings.SplitN(msg.Text, " ", 2)[0][1:]
		return cmd
	}
	return ""
}

// AddCommand modifies the available commands.
//
// If a command already exist this will return an error.
func (c *Client) AddCommand(cmd string, handler Handler) error {
	// Panic if the handler is nil
	if handler == nil {
		return ErrNilHandler
	}
	// Empty command, don't allow
	if len(cmd) == 0 {
		return ErrEmptyCommand
	}
	// If the commands map has not been initialized, do so
	if c.commands == nil {
		c.commands = map[string]Handler{}
	}
	_, ok := c.commands[cmd]

	// Command already exists, return error
	if ok {
		return ErrExistingCommand
	}
	log.Printf("'%s' handler added.", cmd)
	c.commands[cmd] = handler
	return nil
}

// OnShutdown will call fn whenever the process is shutting down
func (c *Client) OnShutdown(fn func() error) {
	c.onShutdown = append(c.onShutdown, fn)
}
