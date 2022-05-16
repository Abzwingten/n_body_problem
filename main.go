package main

import (
	"math"
	"sync"
	// math_rand "math/rand"

	"fmt"
	"os"

	// "sync"

	// "strconv"
	// "strings"
	// "time"

	"github.com/alexflint/go-arg"

	"github.com/ungerik/go3d/float64/vec2"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"

	"n_body_problem/body"
	"n_body_problem/utils"
)

const (
	G         = 6.674e-11
	MinRadius = 4.0
)

var (
	width  = 1000
	height = 900
)

var args struct {
	Dimensions   []int   `arg:"-d,--dimensions" help:"enter window dimensions"`
	Sec_per_tick int     `arg:"-s,--sec_per_tick"`
	Zoom         float64 `arg:"-z,--zoom" default:"1"`
}

type Universe struct {
	scale           float64
	mass_per_planet float64
	sec_per_tick    int
	running         bool
	elapsed         int
	bodies          []*body.Body
	width           int
	height          int
	magnitude       float64
}

func (w Universe) universe_to_screen(coords *vec2.T) vec2.T {
	return vec2.T{coords[0] / w.mass_per_planet * w.scale * w.magnitude, coords[1] / w.mass_per_planet * w.scale * w.magnitude}
}

func (w Universe) universe_time() string {
	d := w.elapsed / (3600 * 24)
	h := (w.elapsed % (3600 * 24)) / 3600
	m := (w.elapsed % 3600) / 60
	s := w.elapsed % 60
	return fmt.Sprintf("%dd %02dh%02dm%02ds", d, h, m, s)
}

func (w Universe) has_escaped(body *body.Body) bool {
	star := w.bodies[0]
	distance := utils.DistanceTo(&body.Position, &star.Position)
	maxDistance := math.Hypot(float64(w.width), float64(w.height)) * 10.0 * w.magnitude * w.mass_per_planet

	return distance > maxDistance && body.Velocity.Length() > math.Sqrt(2.0 * G * star.Mass / distance)
}

func (w *Universe) remove_body(toRemove *body.Body) {
	new_bodies := w.bodies[:0]
	for _, current := range w.bodies {
		if current != toRemove {
			new_bodies = append(new_bodies, current)
		}
	}
	// Clean up remaining
	for i := len(new_bodies); i < len(w.bodies); i++ {
		w.bodies[i] = nil
	}
	w.bodies = new_bodies
}

func (w *Universe) tick() {
	if !w.running {
		return
	}
	for i := 0; i < w.sec_per_tick; i++ {
		w.elapsed += 1
		for _, body := range w.bodies {
			go body.ComputeAcceleration(w.bodies)
		}

		var escaping []*body.Body
		var colliding []map[*body.Body]bool

		// Function literal / closure
		add_collision := func(body1, body2 *body.Body) {
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
			delta_acc := <-body.AccessChannel
			if !math.IsNaN(delta_acc[0]) && !math.IsNaN(delta_acc[1]) {
				// this happens if bodies start out on top of each other
				body.Acceleration[0] = delta_acc[0]
				body.Acceleration[1] = delta_acc[1]
			}
			body.Velocity.Add(&body.Acceleration)
			body.Position.Add(&body.Velocity)

			// Check if body is
			// 1) higher than escape velocity
			// 2) is more more than 2X screens from center.
			if w.has_escaped(body) {
				escaping = append(escaping, body)
			} else {
				for _, body2 := range w.bodies {
					if body == body2 {
						continue
					}
					if body.Collides(body2) {
						add_collision(body, body2)
					}
				}
			}
		}

		// Print escaping
		for _, escapee := range escaping {
			fmt.Printf("%v: ESCAPED: %v\n", w.universe_time(), escapee)
			w.remove_body(escapee)
		}

		for _, group := range colliding {
			var big *body.Body
			for b := range group {
				if big == nil || b.Radius > big.Radius {
					big = b
				}
			}
			for small := range group {
				if small != big {
					big.CollideWith(small)
					fmt.Printf("%v: COLLISION: %v\n", w.universe_time(), big)
					w.remove_body(small)
				}
			}
		}
	}
}

func solarSystem(w, h int) *Universe {
	universe := &Universe{
		scale:           1.0,
		mass_per_planet: 5.5e8,
		sec_per_tick:    600,
		running:         true,
		elapsed:         0,
		bodies:          make([]*body.Body, 6),
		width:           w,
		height:          h,
		magnitude:       1.0,
	}

	sun := body.NewBody("Sol", 0, 0, 696_340_000, 1.9885e30, 0.0, 0.0, 0xFFFF00FF)
	universe.bodies[0] = sun
	universe.bodies[1] = body.NewBody("Mercury", 46e9, 0, 2_439_700, 0.33011e24, 0.0, 58.98e3, 0xAAFF00FF)
	universe.bodies[2] = body.NewBody("Venus", 0, 107.48e9, 6_051_800, 4.86750e24, -35.26e3, 0.0, 0x800000FF)
	universe.bodies[3] = body.NewBody("Mars", 0, -206.62e9, 3_389_500, 0.64171e24, 26.50e3, 0.0, 0xFF0000FF)
	earth := body.NewBody("Earth", -147.09e9, 0, 6_371_000, 5.9724e24, 0.0, -30.29e3, 0x00BBFFFF)
	universe.bodies[4] = earth
	luna := body.NewBody("Luna", earth.Position[0]-0.3633e9, 0, 1_737_400, 0.07346e24, 0.0, earth.Velocity[1]-1.082e3, 0xFFFFFFFF)
	universe.bodies[5] = luna

	// fmt.Printf("BODIES:\n")
	// for _, body := range universe.bodies {
	// 	fmt.Printf("%v\n", body)
	// }
	return universe
}

func run() {

	var universe *Universe

	// Parse argvfont.Close
	arg.MustParse(&args)
	sec_per_tick := args.Sec_per_tick
	magnitude := args.Zoom
	if args.Dimensions != nil {
		width = args.Dimensions[0]
		height = args.Dimensions[1]
	}

	paused := false

	// Init universe
	universe = solarSystem(width, height)
	if width > 0 && height > 0 {
		universe.width = width
		universe.height = height
	}
	if magnitude > 0 {
		universe.magnitude = magnitude
	}
	if sec_per_tick > 0 {
		universe.sec_per_tick = sec_per_tick
	}
	if paused {
		universe.running = false
	}

	// PIXEL INIT part
	cfg := pixelgl.WindowConfig{
		Title:  "Bastinda Space Program",
		Bounds: pixel.R(0, 0, float64(width) * universe.magnitude, float64(height) * universe.magnitude),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// initialize font
	basic_atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	infoTxt := text.New(pixel.V(win.Bounds().Max.X-200*float64(universe.magnitude), win.Bounds().Max.Y-20*float64(universe.magnitude)), basic_atlas)

	follow_body := -1
	center := vec2.T{float64(width / 2), float64(height / 2)}
	offset := center
	var nearest *body.Body

	for !win.Closed() && !win.JustPressed(pixelgl.KeyEscape) {

		// CONTROLS
		if win.JustPressed(pixelgl.KeySpace) {
			universe.running = !universe.running
		}
		// switch center from body to body
		if win.JustPressed(pixelgl.KeyTab) {
			if follow_body == -1 {
				follow_body = 1
			} else {
				follow_body += 1
				if follow_body >= len(universe.bodies) {
					follow_body = 0
				}
			}
		}

		// Recenter
		if win.JustPressed(pixelgl.KeyC) {
			follow_body = -1
			offset = center
		}

		// Turn off closest vec, accel and info display
		if win.JustPressed(pixelgl.MouseButtonRight) {
			nearest = nil
		}

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			nearest = nil
			mouse_coords := win.MousePosition()
			mouse_coords_vec := vec2.T{mouse_coords.X, mouse_coords.Y}
			var nearest_distance float64 = 0.0
			for _, body := range universe.bodies {
				body_screen := universe.universe_to_screen(&body.Position)
				body_screen.Add(&offset)
				if nearest == nil {
					nearest = body
					nearest_distance = utils.DistanceTo(&body_screen, &mouse_coords_vec)
				} else {
					body_distance := utils.DistanceTo(&body_screen, &mouse_coords_vec)
					if body_distance < nearest_distance {
						nearest_distance = body_distance
						nearest = body
					}
				}
			}
		}


		// Speed sim up or down
		if win.Pressed(pixelgl.KeyKPAdd) {
			universe.sec_per_tick += 1
		}
		if win.Pressed(pixelgl.KeyKPSubtract) {
			universe.sec_per_tick -= 1
			if universe.sec_per_tick == 0 {
				universe.sec_per_tick = 1
			}
		}

		// magnitude in/out
		universe.scale *= math.Pow(1.2, win.MouseScroll().Y)

		win.Clear(colornames.Black)

		if follow_body >= 0 && follow_body < len(universe.bodies) {
			offset = vec2.T{center[0], center[1]}
			follow_body_position := universe.universe_to_screen(&universe.bodies[follow_body].Position)
			offset.Sub(&follow_body_position)
		}
		if len(universe.bodies) <= 0 {
			fmt.Println("There are no more bodies, ending sim...")
			os.Exit(3)
		}

		imd := imdraw.New(nil)

		// batch := pixel.NewBatch(&pixel.TrianglesData{}, )
		// Actually drawing
		// in goroutines the expensive stuff
		wait_group := sync.WaitGroup{}
		for i, body := range universe.bodies {
			wait_group.Add(1)
			body_radius_scaled := body.Radius * universe.scale / universe.mass_per_planet
			if body_radius_scaled < MinRadius {
				body_radius_scaled = MinRadius
			}
			body_radius_scaled *= universe.magnitude * universe.scale

			screen_position := universe.universe_to_screen(&body.Position)
			screen_position.Add(&offset)
			go func(i int) {
				planet := universe.bodies[i]
				imd.Color = utils.ColorKeyToColor(planet.Color)
				imd.Push(pixel.V(float64(screen_position[0]), float64(screen_position[1])))
				imd.Circle(float64(body_radius_scaled), 0)

				wait_group.Done()
			}(i)
			wait_group.Wait()
			imd.Draw(win)
		}
		imd.Reset()

		// Update info text
		infoTxt.Clear()
		fmt.Fprintf(infoTxt, "N:\t%v\n", len(universe.bodies))
		fmt.Fprintf(infoTxt, "t:\t%v\n", universe.universe_time())
		fmt.Fprintf(infoTxt, "S:\t%4.2f\n", universe.scale)
		fmt.Fprintf(infoTxt, "dt:\t%v\n", universe.sec_per_tick)

		// Clicked-on body info
		if nearest != nil {
			nearest_position := universe.universe_to_screen(&nearest.Position)
			nearest_position.Add(&offset)

			imd := imdraw.New(nil)
			imd.Color = colornames.Red
			imd.EndShape = imdraw.SharpEndShape

			velocity := nearest.Velocity.Scale(1000)
			end_velocity := nearest_position.Add(velocity)

			// velocity
			imd.Push(pixel.V(nearest_position[0], nearest_position[1]), pixel.V(end_velocity[0], end_velocity[1]))
			imd.Line(2)
			imd.Draw(win)

			// acceleration
			imd.Color = colornames.Green
			acc := nearest.Acceleration.Scale(40)
			endAcc := nearest.Position.Add(acc)
			imd.Push(pixel.V(nearest_position[0], nearest_position[1]), pixel.V(endAcc[0], endAcc[1]))
			imd.Line(2)
			imd.Draw(win)

			fmt.Fprintf(infoTxt, "\n%v:\n", nearest.Name)
			fmt.Fprintf(infoTxt, "P: (%5.2e,%5.2e)\n", nearest.Position[0], nearest.Position[1])
			fmt.Fprintf(infoTxt, "V: (%5.2e,%5.2e)\n", nearest.Velocity[0], nearest.Velocity[1])
			fmt.Fprintf(infoTxt, "A: (%5.2e,%5.2e)\n", nearest.Acceleration[0], nearest.Acceleration[1])
		}
		infoTxt.Draw(win, pixel.IM.Scaled(infoTxt.Orig, universe.magnitude))
		// imd.Reset()
		win.Update()
		universe.tick()
	}
}

func main() {
	pixelgl.Run(run)

}
