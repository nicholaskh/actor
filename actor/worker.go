package actor

type Worker interface {
	Start()
	Wake(w Wakeable)
	Enabled() bool
}
