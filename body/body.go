package body

import (
	"fmt"
	"github.com/Abzwingten/vector.git"
	"math"
)

const G = 6.674e-11

type Body struct {
	Name			string
	Position		vector.Vector
	Velocity		vector.Vector
	Acceleration	vector.Vector
	Radius			float64
	Mass			float62
	AccessChannel	chan vector.Vector
	Color			uint32
}

func NewBody(name string, x float64, y float64, r float64, m float64, vx float64, vy float64, color uint32) *Body {
		return &Body{name, vector.New(x, y), vector.New(vx, vy), vector.New(0, 0), r, m, make(chan vector), color}
}

func NewBodyVector(name string, position vector.Vector, velocity vector.Vector, r float64, m float64, color uint32) *Body {
	return &Body{name, position, velocity, vector.New2DVector(0, 0), r, m, make(chan vector.Vector), color}
}

func (b Body) PrintString() string {
	return fmt.Sprintf("BODY %v: m:%v vel:%v,%v pos:%v,%v r:%v",
		b.Name, b.Mass, b.Velocity.X, b.Velocity.Y, b.Position.X, b.Position.Y, b.Radius)
}

func (b *Body) ComputeAcceleration(other_planets []*Body, c chan vector.Vector) {
	deltaAcc := vector.Vector{0, 0, 0}
	for _, b2 := range planets {
		if b == b2 {
			continue
		}
		distance := math.Hypot(b.Position.X - b2.Position.X, b.Position.Y - b2.Position.Y)
		// distance := math.Sqrt(math.Pow(b.Position.X-b2.Position.X, 2) + math.Pow(b.Position.Y-b2.Position.Y, 2))
		acceleration := vector.New2DVector((b2.Position.X - b.Position.X) / distance, (b2.Position.Y - b.Position.Y) / distance)
		acceleration.MultScalar(G * b2.Mass / (distance * distance))
		deltaAcc.Add(acceleration)
	}
	b.AccessChannel <- deltaAcc
}

func (b Body) Collides(another_planet *Body) bool {
	if &b == another_planet {
		return false
	}
	dx := b.Position.X - another_planet.Position.X
	dy := b.Position.Y - another_planet.Position.Y
	r2 := b.Radius + another_planet.Radius
	return dx*dx+dy*dy-r2*r2 <= 0
}

func (b *Body) CollideWith(another_planet *Body) {
	// Two body problem 
	// Calculate new radius after collision
	// Assume another planet is going away
	new_radius := math.Pow(math.Pow(b.Radius, 3) + math.Pow(another_planet.Radius, 3), 1.0/3.0)
	new_velocity_x := (b.Mass * b.Velocity.X + another_planet.Mass * another_planet.Velocity.X) /
		(b.Mass + another_planet.Mass)
	new_velocity_y := (b.Mass * b.Velocity.Y + another_planet.Mass * another_planet.Velocity.Y) /
		(b.Mass + another_planet.Mass)
	b.Radius = new_radius
	b.Mass += another_planet.Mass
	b.Velocity.X = new_velocity_x
	b.Velocity.Y = new_velocity_y
	b.Name = fmt.Sprintf("%v<-%v", b.Name, another_planet.Name)
}
