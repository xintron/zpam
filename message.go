package zpam

// Message represent outgoing and incoming text events.
//
// A message needs to be able to be used bidirectionally.
type Message struct {
	From string
	To   string
	Text string
}
