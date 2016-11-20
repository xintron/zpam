package irc

import (
	"crypto/tls"
	"fmt"

	irc "github.com/thoj/go-ircevent"
	"github.com/vrischmann/envconfig"
	"github.com/xintron/zpam"
)

func init() {
	zpam.RegisterBackend("irc", newIRC)
}

type ircConfig struct {
	Nickname string
	User     string
	Server   string
	Port     int32

	// Channels should be specified as IRC_CHANNELS="#foo <pass>,#bar"
	Channels []string `envconfig:"optional"`

	// TLS needs to be enabled by setting IRC_TLS=true
	TLS           bool `envconfig:"optional"`
	TLSSkipVerify bool `envconfig:"optional"`
}

type ircBackend struct {
	config     *ircConfig
	connection *irc.Connection
}

// newIRC will create a new Backend connection and parse the IRC configuration
// in the process
func newIRC(c *zpam.Client) zpam.Backend {
	log := zpam.Log.WithField("backend", "irc")
	conf := &ircConfig{}
	err := envconfig.InitWithPrefix(conf, "IRC")
	if err != nil {
		panic(err)
	}

	conn := irc.IRC(conf.Nickname, conf.User)
	conn.UseTLS = conf.TLS
	conn.TLSConfig = &tls.Config{
		ServerName:         conf.Server,
		InsecureSkipVerify: conf.TLSSkipVerify}

	conn.Version = "zpam"

	// Replace the logger with logrus
	w := zpam.Log.Writer()
	conn.Log.SetOutput(w)
	// Make sure to close the logger when the service is shutting down
	c.OnShutdown(func() error {
		return w.Close()
	})
	// Remove the date from the irc logger
	conn.Log.SetFlags(0)

	// Connect will take care of ping-pong for us
	err = conn.Connect(fmt.Sprintf("%s:%d", conf.Server, conf.Port))
	if err != nil {
		panic(err)
	}
	conn.AddCallback("001", func(e *irc.Event) {
		for _, ch := range conf.Channels {
			log.WithField("channel", ch).Debug("Joining IRC channel.")
			conn.Join(ch)
		}
	})
	// Listen for messages and send them to the zpam.Client
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		c.Receive(&zpam.Message{
			From: e.Nick,
			To:   e.Arguments[0],
			Text: e.Message(),
		})
	})
	go conn.Loop()
	return &ircBackend{config: conf, connection: conn}
}

func (i *ircBackend) Send(msg *zpam.Message) {
	i.connection.Privmsg(msg.To, msg.Text)
}
