package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	red   uint32 = 0xffff0000
	blue         = 0xff0000ff
	green        = 0xff00ff00
	black        = 0xff000000
	white        = 0xffffffff
)

var winTitle string = "CitySim"

type TileType int

const (
	_              = iota
	grass TileType = iota
	road
)

type Tile struct {
	Type TileType
}

const GridWidth, GridHeight = 20, 20
const TileSize = 20

const winWidth, winHeight int32 = GridWidth * TileSize, GridHeight * TileSize

type GameGrid [GridWidth][GridHeight]Tile

type ActorType int

const (
	_             = iota
	car ActorType = iota
)

type Actor struct {
	Type  ActorType
	X     int32
	Y     int32
	DestX int32
	DestY int32
}

type Pair struct {
	X int32
	Y int32
}

func MoveCar(a *Actor, state *GameState) bool {
	type pairs struct {
		initial   Pair
		potential Pair
		path      []Pair
	}
	queue := []pairs{}
	dirs := []Pair{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, d := range dirs {
		queue = append(queue, pairs{
			Pair{-1, -1},
			Pair{a.X + d.X, a.Y + d.Y},
			[]Pair{Pair{a.X + d.X, a.Y + d.Y}},
		})
	}
	seen := map[Pair]struct{}{}

	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]

		if _, ok := seen[next.potential]; ok {
			continue
		} else {
			seen[next.potential] = struct{}{}
		}

		if next.potential.X == a.DestX && next.potential.Y == a.DestY &&
			state.Grid[next.potential.X][next.potential.Y].Type == road {
			// Move actor
			if next.initial.X == -1 {
				a.X, a.Y = a.DestX, a.DestY
				return true
			}
			a.X, a.Y = next.initial.X, next.initial.Y
			return true
		}

		if next.potential.X >= 0 && next.potential.X < GridWidth &&
			next.potential.Y >= 0 && next.potential.Y < GridHeight &&
			state.Grid[next.potential.X][next.potential.Y].Type == road {

			var initial Pair
			if next.initial.X == -1 {
				initial = next.potential
			} else {
				initial = next.initial
			}

			for _, d := range dirs {
				queue = append(queue, pairs{
					initial,
					Pair{next.potential.X + d.X, next.potential.Y + d.Y},
					append(next.path[:], next.potential),
				})
			}

		}
	}
	return false
}

func TakeAction(a *Actor, state *GameState) bool {
	var update bool
	switch a.Type {
	case car:
		if a.DestX == -1 ||
			(a.X == a.DestX && a.Y == a.DestY) {
			// Pick new destination
			a.DestX = int32(rand.Intn(GridWidth))
			a.DestY = int32(rand.Intn(GridHeight))

			update = true
		}

		if MoveCar(a, state) {
			update = true
		}
	}
	return update
}

type GameState struct {
	Grid   *GameGrid
	Actors []*Actor
}

func handleMouseButtonEvent(state *GameState, event *sdl.MouseButtonEvent) (bool, error) {
	if event.Type != sdl.MOUSEBUTTONDOWN {
		return false, nil
	}

	var x, y int = int(event.X / TileSize), int(event.Y / TileSize)
	if event.Button == sdl.BUTTON_LEFT {
		state.Grid[x][y].Type = road
	}
	if event.Button == sdl.BUTTON_RIGHT {
		if state.Grid[x][y].Type == road {
			// TODO: Collision
			state.Actors = append(state.Actors, &Actor{car, int32(x), int32(y), -1, -1})
		}
	}
	return true, nil
}

func createState() *GameState {
	grid := GameGrid{}
	for i := 0; i < GridWidth; i++ {
		for j := 0; j < GridHeight; j++ {
			grid[i][j].Type = grass
		}
	}

	state := &GameState{
		Grid:   &grid,
		Actors: []*Actor{},
	}

	return state
}

func drawState(state *GameState, surface *sdl.Surface) {

	// Draw tiles
	for i := int32(0); i < GridWidth; i++ {
		for j := int32(0); j < GridHeight; j++ {
			tileRect := sdl.Rect{i * TileSize, j * TileSize, TileSize, TileSize}

			switch state.Grid[i][j].Type {
			case grass:
				surface.FillRect(&tileRect, green)
			case road:
				surface.FillRect(&tileRect, red)
			}
		}
	}

	// Draw actors
	for _, a := range state.Actors {
		switch a.Type {
		case car:
			// Fill car
			tileRect := sdl.Rect{a.X * TileSize, a.Y * TileSize, TileSize, TileSize}
			surface.FillRect(&tileRect, red)
			carRect := sdl.Rect{
				a.X*TileSize + TileSize/4,
				a.Y*TileSize + TileSize/4,
				TileSize - TileSize/2,
				TileSize - TileSize/2,
			}
			surface.FillRect(&carRect, black)

			// Fill dest
			destRect := sdl.Rect{
				a.DestX*TileSize + TileSize/4,
				a.DestY*TileSize + TileSize/4,
				TileSize - TileSize/2,
				TileSize - TileSize/2,
			}
			surface.FillRect(&destRect, white)

		}
	}
}

func run() error {

	sdl.Init(sdl.INIT_EVERYTHING)

	defer sdl.Quit()

	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		return err
	}

	state := createState()

	drawState(state, surface)
	window.UpdateSurface()

	frameCount := uint32(0)

	fps := uint32(20)

	running := true
	for running {
		frameCount++
		beginTick := sdl.GetTicks()
		var update bool

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quit")
				running = false
			case *sdl.MouseButtonEvent:
				needsUpdate, err := handleMouseButtonEvent(state, e)
				if err != nil {
					fmt.Println(err)
					running = false
					break
				}
				if needsUpdate {
					update = true
				}
			}
		}

		if frameCount%(fps/3) == 0 {

			for _, a := range state.Actors {
				if TakeAction(a, state) {
					update = true
				}
			}
		}

		if update {
			drawState(state, surface)
			window.UpdateSurface()
		}
		endTick := sdl.GetTicks()
		nextTick := beginTick + (1000 / fps)
		if endTick < nextTick {
			sdl.Delay(nextTick - endTick)
		}

	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
