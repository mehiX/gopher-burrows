package burrows

import (
	"log/slog"
	"time"
)

type managedBurrow struct {
	lg *slog.Logger

	requests chan Request
}

// NewManagedBurrow returns a burrow that is managed by a Manager.
// It has its own lifecycle defined in `start()`.
// It owns its data and does not allow direct access to the burrow's data.
func NewManagedBurrow(logger *slog.Logger, initial Burrow) managedBurrow {
	mb := managedBurrow{
		lg:       logger,
		requests: make(chan Request),
	}
	go mb.start(initial)
	return mb
}

func (mb *managedBurrow) start(b Burrow) {

	burrow := b

	pulse := time.NewTicker(Tact)
	defer pulse.Stop()

	for {
		select {
		case <-pulse.C:
			burrow.IncrementAge()
		case req := <-mb.requests:
			switch req.name {
			case ReqClose:
				mb.lg.Info("close burrow", "name", burrow.Name)
				req.response <- Response{burrow: burrow}
				return
			case ReqStatus:
				req.response <- Response{burrow: burrow, nextRequest: nil}
			case ReqAvailable:
				if burrow.IsAvailable() {
					receiveGopher := make(chan Request)
					// send without blocking
					mb.lg.Debug("let the manager know we are available", "name", b.Name)
					select {
					case req.response <- Response{burrow: burrow, nextRequest: receiveGopher}:
						// available so waiting for a new gopher
						select {
						case <-time.After(time.Second):
							// gopher went somewhere else
						case req := <-receiveGopher:
							burrow.Occupied = true
							mb.lg.Debug("sending accept gopher", "name", burrow.Name)
							req.response <- Response{burrow: burrow, nextRequest: nil}
						}
					default:
						mb.lg.Debug("nobody to receive my answer", "name", burrow.Name)
					}

				}
			}
		}
	}
}
