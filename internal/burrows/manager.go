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
	Rentout() (Burrow, error)
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

func (m *manager) Load(in <-chan Burrow) {
	for b := range in {
		m.incoming <- b
	}
}

func (m *manager) List() <-chan managedBurrow {
	all := make(chan managedBurrow)
	m.list <- all
	return all
}

func (m *manager) CurrentStatus() []Burrow {

	ch := make(chan Response)

	req := NewStatusRequest(ch)
	count := 0
	for mb := range m.List() {
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

func (m *manager) Rentout() (Burrow, error) {

	m.lg.Info("start rentout request")

	// ask who is available
	req := NewAvailableRequest()

	m.lg.Debug("send available request to all burrows")
	for mb := range m.List() {
		go func() { mb.requests <- req }()
	}

	select {
	case <-time.After(2 * time.Second):
		return Burrow{}, errors.New("no burrow available")
	case resp := <-req.response:
		gReq := NewGopherRequest()
		resp.nextRequest <- gReq

		waitForResponse := time.NewTimer(time.Second)
		defer waitForResponse.Stop()

		select {
		case <-waitForResponse.C:
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
