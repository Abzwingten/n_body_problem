package main

import(
	"math"
	math_rand "math/rand"

	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/akamensky/argparse"

	"github.com/seifertd/go/vector"
	"n_body_problem/body"
	"n_body_problem/utils"
	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

)


const (
	G         = 6.674e-11
	MinRadius = 4.0
)

type World	struct {
	scale				float64
	mass_per_planet		float64
	sec_per_tick		int
	running				bool
	elapsed				int
	bodies				[]*body.Body
	width				int
	height				int
	zoom				float64
}

func (w World) worldToScreen(coords *vector.Vector) vector.Vector {
	return vector.Vector{coords.X / w.mass_per_planet * w.scale * w.zoom, coords.Y / w.mass_per_planet * w.scale * w.zoom, 0}
}

func (w World) worldTime() string {
	d := w.elapsed / (3600 * 24)
	h := (w.elapsed % (3600 * 24)) / 3600
	m := (w.elapsed % 3600) / 60
	s := w.elapsed % 60
	return fmt.Sprintf("%dd %02dh%02dm%02ds", d, h, m, s)
}



func (w World) has_escaped(body *body.Body) bool {
	star := w.bodies[0]
	radius := body.Pos.DistanceTo(sun.Pos)
	maxDistance := math.Sqrt(float64(IntPow(w.width, 2)+IntPow(w.height, 2))) * 10.0 * w.zoom * w.mass_per_planet

	return radius > maxDistance && body.Vel.Magnitude() > math.Sqrt(2.0 * G * star.Mass / radius)
}


func (w *World) removeBody(toRemove *body.Body) {
	newBodies := w.bodies[:0]
	for _, current := range w.bodies {
		if current != toRemove {
			newBodies = append(newBodies, current)
		}
	}
	// Clean up remaining
	for i := len(newBodies); i < len(w.bodies); i++ {
		w.bodies[i] = nil
	}
	w.bodies = newBodies
}

func (w *World) tick() {
	if !w.running {
		return
	}

	for i := 0; i < w.spt; i++ {
		w.elapsed += 1
		for _, body := range w.bodies {
			go body.CalculateAcceleration(w.bodies)
		}
	
	var escaping []*body.Body
	var colliding []map[*body.Body]bool

	// Function literal / closure
	addCollision := func(body1, body2 *body.Body) {
		fmt.Printf("CRASH! %v AND %v\n", body1.Name, body2.Name)
		added := false
		for _, groups := range colliding {
			if _, ok := groups[body1]; ok {
				groups[body2] = true
				added = true
			}
			if _, ok := groups[body2]; ok {
				groups[body1] = true
				added = true
			}
		}
		if !added {
			newmap := make(map[*body.Body]bool)
			newmap[body1] = true
			newmap[body2] = true
			colliding = append(colliding, newmap)
		}
	}


	for _, body := range w.bodies {
		deltaAcc := <-body.AccessChannel
		if !math.IsNaN(deltaA.X) && !math.IsNaN(deltaA.Y) {
			// this happens if bodies start out on top of each other
			body.Acceleration.X = deltaAcc.X
			body.Acceleration.Y = deltaAcc.Y
		}
		body.Velocity.Add(body.Acceleration)
		body.Position.Add(body.Velocity)

		// Check if body is 1) higher than escape velocity and 2) is more more
		// than 2X screens from center.
		if w.has_escaped(body) {
			escaping = append(escaping, body)
		} else {
			for _, body2 := range w.bodies {
				if body == body2 {
					continue
				}
				if body.Collides(body2) {
					addCollision(body, body2)
				}
			}
		}
	}

	// Print escaping
	for _, escapee := range escaping {
		fmt.Printf("%v: ESCAPED: %v\n", w.worldTime(), escapee)
		w.removeBody(escapee)
	}

	for _, group := range colliding {
		var big *body.Body
		for b, _ := range group {
			if big == nil || b.Radius > big.Radius {
				big = b
			}
		}
		for small, _ := range group {
			if small != big {
				big.CollideWith(small)
				fmt.Printf("%v: COLLISION: %v\n", w.worldTime(), big)
				w.removeBody(small)
			}
		}
	}
}

func solarSystem(w, h int) *World {
	world := &World{
		scale:				1.0,
		mass_per_planet:	5.5e8,
		sec_per_tick:		600,
		running:			true,
		elapsed:			0,
		bodies:				make([]*body.Body, 6),
		width:				w,
		height:				h,
		zoom:				1.0,
	}

	world.bodies[0] = body.NewBody("Sol", 0, 0, 696_340_000, 1.9885e30, 0.0, 0.0, 0xFFFF00)
	world.bodies[1] = body.NewBody("Mercury", 46e9, 0, 2_439_700, 0.33011e24, 0.0, 58.98e3, 0xAAFF00)
	world.bodies[2] = body.NewBody("Venus", 0, 107.48e9, 6_051_800, 4.86750e24, -35.26e3, 0.0, 0x800000)
	world.bodies[3] = body.NewBody("Mars", 0, -206.62e9, 3_389_500, 0.64171e24, 26.50e3, 0.0, 0xFF0000)
	earth			:=  body.NewBody("Earth", -147.09e9, 0, 6_371_000, 5.9724e24, 0.0, -30.29e3, 0x00BBFF)
	world.bodies[4] = earth
	luna			:= body.NewBody("Luna", earth.Pos.X-0.3633e9, 0, 1_737_400, 0.07346e24, 0.0, earth.Vel.Y-1.082e3, 0xC0C0C0)
	world.bodies[5] = luna

	for _, body := range world.bodies {
		fmt.Printf("%v\n", body)
	}
	return world
}

func run() {
	width	:=	800
	height	:=	600

	var world *World
	world = solarSystem(width, height)

	world.zoom = 1

	if sec_per_tick > 0 {
		world.sec_per_tick = sec_per_tick
	}
	if paused {
		world.running = false
	}
}