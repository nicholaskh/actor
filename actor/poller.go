package actor

type Poller interface {
	Poll(ch chan<- Wakeable)
	Stop()
}
