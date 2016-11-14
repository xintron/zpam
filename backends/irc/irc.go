package irc

import (
	"crypto/tls"
	"fmt"

	irc "github.com/thoj/go-ircevent"
	"github.com/xintron/zpam"
)

func init() {
	zpam.RegisterBackend("irc", New)
}

// Config allows for configuration of the IRC bot
var Config = backend{}

type backend struct {
	Nick   string
	User   string
	Server string
	Port   int32
	// Channels should be written as "#foo <password>" where password is
	// optional
	Channels []string
	TLS      *tls.Config

	connection *irc.Connection
}

// New will create a new Backend connection
func New(c *zpam.Client) zpam.Backend {
	conn := irc.IRC(c.Name, c.Name)
	if Config.TLS != nil {
		conn.UseTLS = true
		conn.TLSConfig = Config.TLS
	}

	// Connect will take care of ping-pong for us
	err := conn.Connect(fmt.Sprintf("%s:%d", Config.Server, Config.Port))
	if err != nil {
		panic(err)
	}
	conn.AddCallback("001", func(e *irc.Event) {
		for _, ch := range Config.Channels {
			conn.Join(ch)
		}
	})
	// Listen for messages and send them to the zpam.Client
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		c.Receive(&zpam.Message{
			Text: e.Message(),
		})
	})
	go conn.Loop()
	return &backend{connection: conn}
}

func (i *backend) Send(to string, msg *zpam.Message) {
	i.connection.Privmsg(to, msg.Text)
}
