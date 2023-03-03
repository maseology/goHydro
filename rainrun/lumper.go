package rainrun

// Lumper : interface to lumped rainfall-runoff models
type Lumper interface {
	New(p ...float64)
	Update(p, ep float64) (float64, float64, float64)
	Storage() float64
}
