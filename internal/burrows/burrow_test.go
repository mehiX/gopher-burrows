package burrows

import (
	"math"
	"testing"
)

func TestVolume(t *testing.T) {

	b := Burrow{
		Depth: 2.5,
		Width: 1.2,
	}

	expected := 2.83
	got := b.Volume()

	if math.Abs(expected-got) > 0.01 {
		t.Errorf("wrong volume calculation. expected: %.02f, got: %.02f", expected, got)
	}
}

func TestIncreaseDepth(t *testing.T) {

	scenarios := []struct {
		b     Burrow
		mins  int     // minutes passed
		depth float64 // expected depth
	}{
		{b: Burrow{Name: "new", Depth: 0, Occupied: true}, mins: 1, depth: 0.009},
		{b: Burrow{Name: "new free", Depth: 0, Occupied: false}, mins: 1, depth: 0},
	}

	for _, s := range scenarios {
		t.Run(s.b.Name, func(t *testing.T) {
			t.Parallel()

			for range s.mins {
				s.b.IncrementAge()
			}

			if math.Abs(s.b.Depth-s.depth) > 0.001 {
				t.Errorf("wrong depth after %d mins. expected: %.3f, got: %.3f", s.mins, s.depth, s.b.Depth)
			}
		})
	}
}

func TestIsAvailable(t *testing.T) {

	scenarios := []struct {
		b         Burrow
		available bool
	}{
		{b: Burrow{Name: "one min to collapse free", AgeInMin: maxAgeInMin - 1, Occupied: false}, available: true},
		{b: Burrow{Name: "one min to collapse occupied", AgeInMin: maxAgeInMin - 1, Occupied: true}, available: false},
		{b: Burrow{Name: "collapsing", AgeInMin: maxAgeInMin}, available: false},
		{b: Burrow{Name: "just collapsed", AgeInMin: maxAgeInMin + 1}, available: false},
		{b: Burrow{Name: "long collapsed", AgeInMin: maxAgeInMin + 100}, available: false},
		{b: Burrow{Name: "good free", AgeInMin: 19, Occupied: false}, available: true},
		{b: Burrow{Name: "good occupied", AgeInMin: 19, Occupied: true}, available: false},
	}

	for _, s := range scenarios {
		t.Run(s.b.Name, func(t *testing.T) {
			t.Parallel()

			av := s.b.IsAvailable()

			if av != s.available {
				t.Errorf("wrong availability. expected: %v, got: %v", s.available, av)
			}
		})
	}
}

func TestIncreaseAge(t *testing.T) {

	scenarios := []struct {
		b         *Burrow
		mins      int // minutes passed
		available bool
	}{
		{b: &Burrow{Name: "collapsed free", Occupied: false, AgeInMin: 10}, mins: maxAgeInMin - 9, available: false},
		{b: &Burrow{Name: "collapsed occupied", Occupied: true, AgeInMin: 10}, mins: maxAgeInMin - 9, available: false},
		{b: &Burrow{Name: "not collapsed free", Occupied: false, AgeInMin: 10}, mins: maxAgeInMin - 11, available: true},
		{b: &Burrow{Name: "not collapsed occupied", Occupied: true, AgeInMin: 10}, mins: maxAgeInMin - 11, available: false},
	}

	for _, s := range scenarios {
		t.Run(s.b.Name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < s.mins; i++ {
				s.b.IncrementAge()
			}

			av := s.b.IsAvailable()

			if av != s.available {
				t.Errorf("wrong availability after time passed. Age is: %v", s.b.AgeInMin)
			}
		})
	}
}
