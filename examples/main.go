package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/xintron/zpam"
	_ "github.com/xintron/zpam/backends/hipchat"
	_ "github.com/xintron/zpam/backends/irc"
	"gopkg.in/alecthomas/kingpin.v2"
)

var ()

var TableFlipper = &tableFlipper{
	Tables: []string{
		"(╯°□°）╯︵ ┻━┻",
		"(┛◉Д◉)┛彡┻━┻",
		"(ﾉ≧∇≦)ﾉ ﾐ ┸━┸",
		"(ノಠ益ಠ)ノ彡┻━┻",
		"(╯ರ ~ ರ）╯︵ ┻━┻",
		"(┛ಸ_ಸ)┛彡┻━┻",
		"(ﾉ´･ω･)ﾉ ﾐ ┸━┸",
		"(ノಥ,_｣ಥ)ノ彡┻━┻",
		"(┛✧Д✧))┛彡┻━┻",
	},
	Rooms: map[string]*time.Ticker{},
}

type tableFlipper struct {
	Tables        []string
	Rooms         map[string]*time.Ticker
	log           *logrus.Entry
	parser        *kingpin.Application
	startInterval *time.Duration
}

func (t *tableFlipper) init() {
	t.parser = kingpin.New("table flipper", "We flip tables!")

	t.parser.Command("flip", "Flipp table.").Default()
	t.parser.HelpCommand = nil
	t.parser.HelpFlag = nil

	start := t.parser.Command("start", "Start timer to flip tables on an interval.")
	t.startInterval = start.Flag("interval", "Time between flips in seconds.").Short('i').Default("1m").Duration()

	t.parser.Command("stop", "Stop running timer.")

	t.log = zpam.Log.WithField("plugin", "tableflipper")
}

func (t *tableFlipper) Handle(c *zpam.Client, msg *zpam.Message) {
	// Parse the incoming data with kingpin
	args := strings.Split(msg.Text, " ")[1:]
	cmd, err := t.parser.Parse(args)
	if err != nil {
		c.Send(&zpam.Message{To: msg.To, Text: "Invalid command: " + err.Error()})
		return
	}

	switch cmd {
	case t.parser.GetCommand("start").FullCommand():
		// Check if there already is a flipper running, if so stop it and start
		// the new one
		flipper, ok := t.Rooms[msg.To]
		if ok {
			flipper.Stop()
			delete(t.Rooms, msg.To)
		}

		t.log.WithField("interval", t.startInterval).Debug("starting ticker.")
		ticker := time.NewTicker(*t.startInterval)
		// Start the background worker
		go func(c *zpam.Client, ch <-chan time.Time, room string) {
			for _ = range ch {
				c.Send(&zpam.Message{
					To:   room,
					Text: t.Tables[rand.Intn(len(t.Tables))],
				})
			}
		}(c, ticker.C, msg.To)
		t.Rooms[msg.To] = ticker

		c.Send(&zpam.Message{
			To:   msg.To,
			Text: "I will now flipp tables every " + t.startInterval.String(),
		})

	case t.parser.GetCommand("stop").FullCommand():
		flipper, ok := t.Rooms[msg.To]
		if ok {
			flipper.Stop()
			delete(t.Rooms, msg.To)
			c.Send(&zpam.Message{
				To:   msg.To,
				Text: "I'm calm again, no more table flipping",
			})
		}
	default:
		c.Send(&zpam.Message{
			To:   msg.To,
			Text: t.Tables[rand.Intn(len(t.Tables))],
		})
	}

}

func main() {
	client := zpam.Client{}

	client.AddCommand("time", zpam.HandlerFunc(func(c *zpam.Client, msg *zpam.Message) {
		zpam.Log.WithField("msg", msg).Debug("What time is it?")
		c.Send(&zpam.Message{To: msg.To, Text: time.Now().Format(time.RFC3339)})
	}))

	TableFlipper.init()
	client.AddCommand("table", TableFlipper)

	rand.Seed(time.Now().UnixNano())
	err := client.Run()
	if err != nil {
		panic(err)
	}
}
