package zpam

// Backend interface that backends need to implement
//
// All bootstraping and starting should take place within the Setup function.
// To handle graceful shutdowns the Backend can (and probably should) register
// a shutdown function with *Client.OnShutdown()
type Backend interface {
	Send(msg *Message)
}

type initiator func(*Client) Backend

var backends = map[string]initiator{}

// RegisterBackend supports multiple backends at the same time.
//
// Even though only one backend can run per process, multiple can be
// registered. This allows one to build a binary with multiple backends ready
// and then control which backend to use in the configuration file.
func RegisterBackend(name string, init initiator) {
	backends[name] = init
	Log.WithField("backend", name).Debug("registered backend")
}
