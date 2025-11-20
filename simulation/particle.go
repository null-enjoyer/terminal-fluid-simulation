package simulation

type Vector struct {
	X, Y float64
}

type Particle struct {
	Pos    Vector
	OldPos Vector
}
