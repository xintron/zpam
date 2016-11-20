package backends

import (
	"strings"

	"github.com/daneharrigan/hipchat"
	"github.com/vrischmann/envconfig"
	"github.com/xintron/zpam"
)

func init() {
	zpam.RegisterBackend("hipchat", newHipchat)
}

type hcConfig struct {
	Name     string
	User     string
	Password string
	Rooms    []string
}

type hcBackend struct {
	config     *hcConfig
	connection *hipchat.Client
}

// New will create a new Backend connection
func newHipchat(c *zpam.Client) zpam.Backend {
	log := zpam.Log.WithField("backend", "hipchat")
	conf := &hcConfig{}
	err := envconfig.InitWithPrefix(conf, "HIPCHAT")
	if err != nil {
		panic(err)
	}

	client, err := hipchat.NewClient(conf.User, conf.Password, "bot")
	if err != nil {
		panic(err)
	}
	b := &hcBackend{
		config:     conf,
		connection: client,
	}

	client.Status("chat")
	for _, room := range conf.Rooms {
		client.Join(room, conf.Name)
		log.WithField("room", room).Debug("joining room.")
	}

	// Start listening for incoming messages and send them to the zpam.Client
	go func(in <-chan *hipchat.Message) {
		for {
			msg := <-in
			c.Receive(&zpam.Message{
				From: msg.From,
				To:   strings.SplitN(msg.From, "/", 2)[0],
				Text: msg.Body,
			})
		}
	}(client.Messages())

	// Keep-alive will send a message to hipchat every 60 second. If no data
	// has been sent the client will be disconnected after 150 seconds from the
	// hipchat server.
	go client.KeepAlive()
	return b
}

func (b *hcBackend) Send(msg *zpam.Message) {
	b.connection.Say(msg.To, b.config.Name, msg.Text)
}
