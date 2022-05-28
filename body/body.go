package body

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/barnex/fmath"
)

const G = 6.674e-11

type Body struct {
	Name          string
	Position      rl.Vector2
	Velocity      rl.Vector2
	Acceleration  rl.Vector2
	Radius        float32
	Mass          float32
	AccessChannel chan rl.Vector2
	Color         uint32
}

func NewBody(name string, x float32, y float32, r float32, m float32, vx float32, vy float32, color uint32) *Body {
	return &Body{name, rl.Vector2{X: x, Y: y}, rl.Vector2{X: vx, Y: vy}, rl.Vector2Zero(), r, m, make(chan rl.Vector2), color}
}

func NewBodyVector(name string, position rl.Vector2, velocity rl.Vector2, r float32, m float32, color uint32) *Body {
	return &Body{name, position, velocity, rl.Vector2Zero(), r, m, make(chan rl.Vector2), color}
}

func (b Body) Print_body() string {
	return fmt.Sprintf("BODY %v: m:%v vel:%v,%v pos:%v,%v r:%v",
		b.Name, b.Mass, b.Velocity.X, b.Velocity.Y, b.Position.X, b.Position.Y, b.Radius)
}

func (b *Body) ComputeAcceleration(other_planets []*Body) {
	delta_acc := rl.Vector2Zero()
	for _, b2 := range other_planets {
		if b == b2 {
			continue
		}
		distance := fmath.Hypot(b.Position.X-b2.Position.X, b.Position.Y-b2.Position.Y)
		acceleration := rl.Vector2{X: (b2.Position.X - b.Position.X) / distance, Y: (b2.Position.Y - b.Position.Y) / distance}
		acceleration = rl.Vector2Scale(acceleration, G * b2.Mass / (fmath.Pow(distance, 2)))
		delta_acc = rl.Vector2Add(delta_acc, acceleration)
	}
	b.AccessChannel <- delta_acc
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
	new_radius := fmath.Pow(fmath.Pow(b.Radius, 3)+fmath.Pow(another_planet.Radius, 3), 1.0/3.0)
	new_velocity_x := (b.Mass * b.Velocity.X + another_planet.Mass * another_planet.Velocity.Y) /
		(b.Mass + another_planet.Mass)
	new_velocity_y := (b.Mass*b.Velocity.X + another_planet.Mass * another_planet.Velocity.Y) /
		(b.Mass + another_planet.Mass)
	b.Radius = new_radius
	b.Mass += another_planet.Mass
	b.Velocity.X = new_velocity_x
	b.Velocity.Y = new_velocity_y
	b.Name = fmt.Sprintf("%v<-%v", b.Name, another_planet.Name)
}
