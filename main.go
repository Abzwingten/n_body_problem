package main

import (
	"math"
	// math_rand "math/rand"

	"fmt"
	"os"
	"sync"

	// "strconv"
	// "strings"
	// "time"

	"github.com/alexflint/go-arg"

	"github.com/ungerik/go3d/fmath"
	"github.com/ungerik/go3d/vec2"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"golang.org/x/image/colornames"

	"n_body_problem/body"
	"n_body_problem/utils"
)

const (
	G         = 6.674e-11
	MinRadius = 4.0
)

const (
	frame_rate = 60
	font_path  = "assets/Futura_book.ttf"
	font_size  = 16
)

var (
	width		=	1920
	height		=	1050
)

var args struct {
	Dimensions   []int   `arg:"-d,--dimensions" help:"enter window dimensions"`
	Sec_per_tick int     `arg:"-s,--sec_per_tick"`
	Zoom         float32 `arg:"-z,--zoom" default:"1"`
}

type Universe struct {
	scale           float32
	mass_per_planet float32
	sec_per_tick    int
	running         bool
	elapsed         int
	bodies          []*body.Body
	width           int
	height          int
	zoom            float32
}

func (w Universe) universe_to_screen(coords *vec2.T) vec2.T {
	return vec2.T{coords[0] / w.mass_per_planet * w.scale * w.zoom, coords[1] / w.mass_per_planet * w.scale * w.zoom}
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
	maxDistance := fmath.Hypot(float32(w.width), float32(w.height)) * 10.0 * w.zoom * w.mass_per_planet

	return distance > maxDistance && body.Velocity.Length() > fmath.Sqrt(2.0 * G * star.Mass / distance)
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
			deltaAcc := <-body.AccessChannel
			if !math.IsNaN(float64(deltaAcc[0])) && !math.IsNaN(float64(deltaAcc[1])) {
				// this happens if bodies start out on top of each other
				body.Acceleration[0] = deltaAcc[0]
				body.Acceleration[1] = deltaAcc[1]
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
		mass_per_planet: 5.5e6,
		sec_per_tick:    600,
		running:         true,
		elapsed:         0,
		bodies:          make([]*body.Body, 6),
		width:           w,
		height:          h,
		zoom:            1.0,
	}

	sun := body.NewBody("Sol", 0, 0, 696_340_000, 1.9885e30, 0.0, 0.0, 0xFFFF00FF)
	universe.bodies[0] = sun
	universe.bodies[1] = body.NewBody("Mercury", 46e9, 0, 2_439_700, 0.33011e24, 0.0, 58.98e3, 0xAAFF00FF)
	universe.bodies[2] = body.NewBody("Venus", 0, 107.48e9, 6_051_800, 4.86750e24, -35.26e3, 0.0, 0x800000FF)
	universe.bodies[3] = body.NewBody("Mars", 0, -206.62e9, 3_389_500, 0.64171e24, 26.50e3, 0.0, 0xFF0000FF)
	earth := body.NewBody("Earth", -147.09e9, 0, 6_371_000, 5.9724e24, 0.0, -30.29e3, 0x00BBFFFF)
	universe.bodies[4] = earth
	luna := body.NewBody("Luna", earth.Position[0]-0.3633e9, 0, 1_737_400, 0.07346e24, 0.0, earth.Velocity[1]-1.082e3, 0xC0C0C0FF)
	universe.bodies[5] = luna

	// fmt.Printf("BODIES:\n")
	// for _, body := range universe.bodies {
	// 	fmt.Printf("%v\n", body)
	// }
	return universe
}

func run() int {

	var universe *Universe

	// Parse argvfont.Close
	arg.MustParse(&args)
	sec_per_tick := args.Sec_per_tick
	zoom := args.Zoom
	if args.Dimensions != nil {
		width	=	args.Dimensions[0]
		height	=	args.Dimensions[1]
	}

	paused := false

	// Init universe
	universe = solarSystem(width, height)
	if width > 0 && height > 0 {
		universe.width = width
		universe.height = height
	}
	if zoom > 0 {
		universe.zoom = zoom
	}
	if sec_per_tick > 0 {
		universe.sec_per_tick = sec_per_tick
	}
	if paused {
		universe.running = false
	}

	// SDL INIT part

	var window *sdl.Window
	var renderer *sdl.Renderer
	// var texture *sdl.Texture
	var text *sdl.Surface
	var err error

	var running_mutex sync.Mutex

	// SDL
	sdl.Do(func() {
		if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize SDL: %s\n", err)
		}
	})
	defer sdl.Quit()

	// WINDOW
	sdl.Do(func() {
		window, err = sdl.CreateWindow("Bastinda Space Program", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(width), int32(height), sdl.WINDOW_SHOWN)
	})
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		sdl.Do(func() {
			window.Destroy()
		})
	}()

	// RENDERER
	sdl.Do(func() {
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	})
	if err != nil {
		fmt.Println(err)
	}
	sdl.Do(func() {
		renderer.Clear()
	})
	defer func() {
		sdl.Do(func() {
			renderer.Destroy()
		})
	}()

	if err = ttf.Init(); err != nil {
		return 1
	}
	defer ttf.Quit()

	font, err := ttf.OpenFont(font_path, font_size)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	defer func() {
		sdl.Do(func() {
			font.Close()

		})
	}()

	follow_body := -1
	center := vec2.T{float32(width / 2), float32(height / 2)}
	offset := center
	var nearest *body.Body

	get_planet_info := func(x, y int32) {
		nearest = nil
		mouse_coords_vec := vec2.T{float32(x), float32(y)}
		var nearest_distance float32 = 0.0
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

	running := true
	for running {
		// CONTROLS
		sdl.Do(func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					running_mutex.Lock()
					running = false
					running_mutex.Unlock()
				case *sdl.KeyboardEvent:
					switch t.Keysym.Sym {
					case sdl.K_ESCAPE:
						running_mutex.Lock()
						running = false
						running_mutex.Unlock()
					case sdl.K_SPACE:
						universe.running = !universe.running
					case sdl.K_TAB:
						if follow_body == -1 {
							follow_body = 1
						} else {
							follow_body += 1
							if follow_body >= len(universe.bodies) {
								follow_body = 0
							}
						}
					case sdl.K_c:
						follow_body = -1
						offset = center
					case sdl.K_KP_PLUS:
						universe.sec_per_tick += 1
						fmt.Println(universe.sec_per_tick)
					case sdl.K_KP_MINUS:
						universe.sec_per_tick -= 1
					}

				case *sdl.MouseButtonEvent:
					switch t.Button {
					case sdl.BUTTON_RIGHT:
						nearest = nil
					case sdl.BUTTON_LEFT:
						get_planet_info(t.X, t.Y)
					}
				case *sdl.MouseWheelEvent:
					universe.zoom *= fmath.Pow(1.2, float32(t.Y))
				}

				if follow_body >= 0 && follow_body < len(universe.bodies) {
					offset = vec2.T{center[0], center[1]}
					follow_body_position := universe.universe_to_screen(&universe.bodies[follow_body].Position)
					offset.Sub(&follow_body_position)
				}
				if len(universe.bodies) <= 0 {
					fmt.Println("No bodies left, ending simulation")
					os.Exit(3)
				}
			}
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.FillRect(&sdl.Rect{0, 0, int32(width), int32(height)})
		})

		// Actually drawing
		// in goroutines thn expensive stuff
		wait_group := sync.WaitGroup{}
		for i, body := range universe.bodies {
			wait_group.Add(1)

			body_radius_scaled := body.Radius * universe.scale / universe.mass_per_planet
			if body_radius_scaled < MinRadius {
				body_radius_scaled = MinRadius
			}
			body_radius_scaled *= universe.zoom * universe.scale

			screen_position := universe.universe_to_screen(&body.Position)
			screen_position.Add(&offset)
			// planet := universe.bodies[i]

			// gfx.CircleColor(renderer, int32(screen_position[0] + body_radius_scaled / 4), int32(screen_position[1] + body_radius_scaled / 4), int32(body_radius_scaled), utils.ColorKeyToSDL(planet.Color))

			go func(i int) {
				planet := universe.bodies[i]
				sdl.Do(func() {
					// gfx.CharacterColor(renderer, int32(screen_position[0]), int32(screen_position[1]), '*', sdl.Color{255, 0, 0, 255})
					// gfx.CircleColor(renderer, 200, 200, 50, utils.ColorKeyToSDL(planet.Color))
					gfx.FilledCircleColor(renderer, int32(screen_position[0] + body_radius_scaled / 4), int32(screen_position[1] + body_radius_scaled / 4), int32(body_radius_scaled), utils.ColorKeyToSDL(planet.Color))
				})
				// fmt.Println(int32(body_radius_scaled))
				wait_group.Done()
			}(i)
		}
		wait_group.Wait()

		// Clicked-on body info
		if nearest != nil {
			nearest_position := universe.universe_to_screen(&nearest.Position)
			nearest_position.Add(&offset)
			velocity := nearest.Velocity.Scale(1000)
			end_velocity := nearest_position.Add(velocity)

			gfx.ThickLineColor(renderer, int32(nearest_position[0]), int32(nearest_position[1]), int32(end_velocity[0]), int32(end_velocity[1]), 20, sdl.Color(colornames.Aquamarine))
		}
		universe.universe_time()

		sdl.Do(func() {
			renderer.Present()
			sdl.Delay(1000 / frame_rate)
		})

		sdl.Do(func() {
			universe.tick()
		})
	}
	return 0
}

func main() {
	// os.Exit(..) must run AFTER sdl.Main(..) below; so keep track of exit
	// status manually outside the closure passed into sdl.Main(..) below
	var exitcode int
	sdl.Main(func() {
		exitcode = run()
	})
	// os.Exit(..) must run here! If run in sdl.Main(..) above, it will cause
	// premature quitting of sdl.Main(..) function; resource cleaning deferred
	// calls/closing of channels may never run
	os.Exit(exitcode)
}
