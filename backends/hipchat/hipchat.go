package hipchat

import (
	"github.com/daneharrigan/hipchat"
	"github.com/xintron/zpam"
)

func init() {
	zpam.RegisterBackend("hipchat", New)
}

// Config allows the bot to be configured. This should be done before New is
// called.
var Config = backend{}

type backend struct {
	Name       string
	User       string
	Password   string
	Rooms      []string
	connection *hipchat.Client
}

// New will create a new Backend connection
func New(c *zpam.Client) zpam.Backend {
	client, err := hipchat.NewClient(Config.User, Config.Password, "bot")
	if err != nil {
		panic(err)
	}
	Config.connection = client

	client.Status("chat")
	for _, room := range Config.Rooms {
		client.Join(room, Config.Name)
	}

	// Start listening for incoming messages and send them to the zpam.Client
	go func(in <-chan *hipchat.Message) {
		for {
			msg := <-in
			c.Receive(&zpam.Message{
				Text: msg.Body,
			})
		}
	}(client.Messages())
	Config.connection = client
	return &Config
}

func (b *backend) Send(to string, msg *zpam.Message) {
	b.connection.Say(to, b.Name, msg.Text)
}
