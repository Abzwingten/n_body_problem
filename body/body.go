package body

import (
	"fmt"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/fmath"
)

const G = 6.674e-11

type Body struct {
	Name			string
	Position		vec2.T
	Velocity		vec2.T
	Acceleration	vec2.T
	Radius			float32
	Mass			float32
	AccessChannel	chan vec2.T
	Color			uint32
}

func NewBody(name string, x float32, y float32, r float32, m float32, vx float32, vy float32, color uint32) *Body {
		return &Body{name, vec2.T{x, y}, vec2.T{vx, vy}, vec2.T{0, 0}, r, m, make(chan vec2.T), color}
}

func NewBodyVector(name string, position vec2.T, velocity vec2.T, r float32, m float32, color uint32) *Body {
	return &Body{name, position, velocity, vec2.T{0, 0}, r, m, make(chan vec2.T), color}
}

func (b Body) Print_body() string {
	return fmt.Sprintf("BODY %v: m:%v vel:%v,%v pos:%v,%v r:%v",
		b.Name, b.Mass, b.Velocity[0], b.Velocity[1], b.Position[0], b.Position[1], b.Radius)
}

func (b *Body) ComputeAcceleration(other_planets []*Body) {
	deltaAcc := vec2.T{0, 0}
	for _, b2 := range other_planets {
		if b == b2 {
			continue
		}
		distance := fmath.Hypot(b.Position[0] - b2.Position[0], b.Position[1] - b2.Position[1])
		// distance := float32(math.Sqrt(math.Pow(float64(b.Position[0] - b2.Position[1]), 2) + math.Pow(float64(b.Position[0] - b2.Position[1]), 2)))
		acceleration := vec2.T{(b2.Position[0] - b.Position[0]) / distance, (b2.Position[1] - b.Position[1]) / distance}
		acceleration.Scale(G * b2.Mass / (distance * distance))
		deltaAcc.Add(&acceleration)
	}
	b.AccessChannel <- deltaAcc
}

func (b Body) Collides(another_planet *Body) bool {
	if &b == another_planet {
		return false
	}
	dx := b.Position[0] - another_planet.Position[0]
	dy := b.Position[1] - another_planet.Position[1]
	r2 := b.Radius + another_planet.Radius
	return dx*dx+dy*dy-r2*r2 <= 0
}

func (b *Body) CollideWith(another_planet *Body) {
	// Two body problem
	// Calculate new radius after collision
	// Assume another planet is going away
	new_radius := fmath.Pow(fmath.Pow(b.Radius, 3) + fmath.Pow(another_planet.Radius, 3), 1.0/3.0)
	new_velocity_x := (b.Mass * b.Velocity[0] + another_planet.Mass * another_planet.Velocity[0]) /
		(b.Mass + another_planet.Mass)
	new_velocity_y := (b.Mass * b.Velocity[1] + another_planet.Mass * another_planet.Velocity[1]) /
		(b.Mass + another_planet.Mass)
	b.Radius = new_radius
	b.Mass += another_planet.Mass
	b.Velocity[0] = new_velocity_x
	b.Velocity[1] = new_velocity_y
	b.Name = fmt.Sprintf("%v<-%v", b.Name, another_planet.Name)
}
