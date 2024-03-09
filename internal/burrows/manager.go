package burrows

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// tact is used for testing in order to make the time go faster. Normally it should be set to 1 minute.
var tact = 3 * time.Second

type Report struct {
	TotalDepth    float64
	NumAvailable  int
	VolumeMin     float64
	VolumeMinName string
	VolumeMax     float64
	VolumeMaxName string
}

type Manager interface {
	Load(<-chan Burrow)
	CurrentStatus() []Burrow
	Rentout(ctx context.Context) (Burrow, error)
	Report() Report
}

type manager struct {
	lg *slog.Logger

	// only internal. should not be accessed directly. use the list channel
	burrows []managedBurrow

	// list receives requests to expose the list of managedBurrows
	list chan chan managedBurrow

	incoming chan Burrow

	// Done will be closed by the manager once all cleanup is done
	Done chan struct{}
}

// NewManager creates a new burrows manager.
// It starts a go routine that manages the lifecycle of the manager
func NewManager(ctx context.Context, logger *slog.Logger) *manager {
	m := &manager{
		lg:       logger,
		list:     make(chan chan managedBurrow),
		incoming: make(chan Burrow),
		Done:     make(chan struct{}),
	}
	go m.manage(ctx)
	return m
}

func (m *manager) manage(ctx context.Context) {
	m.lg.Debug("start manage")

	defer close(m.Done)

	for {
		select {
		case <-ctx.Done():
			m.lg.Info("received closing signal", "service", "manager")
			// Save data if needed
			for _, b := range m.burrows {
				b.requests <- Request{name: ReqClose}
			}
			return
		case b := <-m.incoming:
			managedBurrow := NewManagedBurrow(m.lg, b)
			m.burrows = append(m.burrows, managedBurrow)
			m.lg.Info("managing new burrow", "name", b.Name)
		case lst := <-m.list:
			go func() {
				defer close(lst)
				for _, b := range m.burrows {
					select {
					case <-ctx.Done():
						return
					case lst <- b:
					}
				}
			}()
		}
	}

}

// Load reads data from the incoming channel and stores it in the internal structure of the manager.
// It is safe to call `Load` in a separate go routine
func (m *manager) Load(in <-chan Burrow) {
	for b := range in {
		m.incoming <- b
	}
}

// CurrentStatus returns a list of all the burrows currently managed.
func (m *manager) CurrentStatus() []Burrow {

	ch := make(chan Response)

	req := NewStatusRequest(ch)
	count := 0
	for mb := range m.stream() {
		count++
		go func() { mb.requests <- req }()
	}

	burrows := make([]Burrow, count)
	for i := 0; i < count; i++ {
		resp := <-ch
		burrows[i] = resp.burrow
	}

	return burrows
}

// Rentout picks the first available burrow and assigns it to a gopher by returning it to the caller.
// If no available burrow can be found then an error is returned.
// The passed in context can control how long the renting process can last. It returns an error if
// the context expires before a burrow could be rented out.
func (m *manager) Rentout(ctx context.Context) (Burrow, error) {

	m.lg.Info("start rentout request")

	// ask who is available
	req := NewAvailableRequest()

	m.lg.Debug("send available request to all burrows")
	for mb := range m.stream() {
		go func() { mb.requests <- req }()
	}

	select {
	case <-ctx.Done():
		return Burrow{}, errors.New("no burrow available")
	case resp := <-req.response:
		gReq := NewGopherRequest()
		resp.nextRequest <- gReq

		select {
		case <-ctx.Done():
			return Burrow{}, errors.New("available burrow did not respond in time")
		case resp := <-gReq.response:
			return resp.burrow, nil
		}
	}
}

func (m *manager) Report() Report {

	rep := Report{}

	burrows := m.CurrentStatus()

	for _, b := range burrows {
		rep.TotalDepth += b.Depth

		if b.IsAvailable() {
			rep.NumAvailable++
		}

		vol := b.Volume()
		if rep.VolumeMin == 0 || vol < rep.VolumeMin {
			rep.VolumeMin = vol
			rep.VolumeMinName = b.Name
		}

		if rep.VolumeMax < vol {
			rep.VolumeMax = vol
			rep.VolumeMaxName = b.Name
		}
	}

	return rep
}

// stream returns a channel where it sends all the burrows that
// the manager manages at the moment.
// It is thread safe and meant to be used internally to expose data to other go routines.
func (m *manager) stream() <-chan managedBurrow {
	all := make(chan managedBurrow)
	m.list <- all
	return all
}
