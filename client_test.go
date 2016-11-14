package zpam

import (
	"sync/atomic"
	"testing"
)

func TestClientCommandHandler(t *testing.T) {
	var handleCount int32
	var h HandlerFunc = func(*Client, *Message) {
		atomic.AddInt32(&handleCount, 1)
	}

	client := &Client{
		Name:   "zpam",
		Prefix: ".",
	}
	client.AddCommand("test", h)

	client.Receive(&Message{
		Text: ".test this should work",
	})
	// This will should not trigger the command handler
	client.Receive(&Message{
		Text: "Normal message without command trigger",
	})
	if handleCount != 1 {
		t.Errorf("expected handler to run one time, ran %d times\n", handleCount)
	}
}

func TestClientParseCommand(t *testing.T) {
	c := &Client{
		Prefix: ".",
	}
	cmd := c.parseCommand(&Message{Text: ".test"})
	if cmd != "test" {
		t.Errorf("expected command to be 'test', actual value '%s'", cmd)
	}

	emptyCmds := []Message{
		Message{Text: "foobar"},
		Message{Text: ". lol"},
		Message{Text: ""},
		Message{Text: "&.no test"},
		Message{Text: "foo .no bar"},
	}
	for _, msg := range emptyCmds {
		cmd = c.parseCommand(&msg)
		if cmd != "" {
			t.Errorf("expected command to be an empty string, actual value '%s'", cmd)
		}
	}
}

func TestClientAddCommands(t *testing.T) {
	var h HandlerFunc = func(*Client, *Message) {}
	client := &Client{}
	client.AddCommand("test", h)
	err := client.AddCommand("test", h)

	if err != ErrExistingCommand {
		t.Error("Same command added twice. Expected an error the second time.")
	}
}
