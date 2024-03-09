package burrows

import "math"

const maxAgeInMin int = 25 * 24 * 60 // 25 days

type Burrow struct {
	Name     string  `json:"name"`
	Occupied bool    `json:"occupied"`
	Depth    float64 `json:"depth"`
	Width    float64 `json:"width"`
	AgeInMin int     `json:"age"`
}

// IsAvailable returns `true` if the burrow is not occupied by a gopher and if it hasn't already collapsed.
// A burrow collapses automatically after exactly 25 days
func (b *Burrow) IsAvailable() bool {
	return !b.Occupied && b.AgeInMin < maxAgeInMin
}

// Volume returns the volume of the burrow.
// The burrow has a cylindrical shape with known depth and radius.
func (b *Burrow) Volume() float64 {
	return b.Depth * math.Pi * math.Pow(b.Width, 2) / 4
}

func (b *Burrow) IncrementAge() {
	if b.AgeInMin+1 > maxAgeInMin {
		return
	}

	b.AgeInMin++
	if b.Occupied {
		if b.Depth == 0.0 {
			b.Depth = 0.009
		} else {
			b.Depth += b.Depth * 0.009
		}
	}
}
