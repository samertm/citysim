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

const GridWidth, GridHeight = 80, 60
const TileSize = 20

const winWidth, winHeight int32 = GridWidth * TileSize, GridHeight * TileSize

type GameGrid [GridWidth][GridHeight]Tile

type ActorType int

const (
	_             = iota
	car ActorType = iota
)

type Actor struct {
	Type ActorType
	X    int32
	Y    int32
}

func TakeAction(a *Actor, state *GameState) bool {
	var update bool
	switch a.Type {
	case car:
		dir := rand.Intn(4)
		switch dir {
		case 0: // up
			if a.Y-1 >= 0 &&
				state.Grid[a.X][a.Y-1].Type == road {
				a.Y--
				update = true
			}
		case 1: // down
			if a.Y+1 < GridHeight &&
				state.Grid[a.X][a.Y+1].Type == road {
				a.Y++
				update = true
			}
		case 2: // left
			if a.X-1 >= 0 &&
				state.Grid[a.X-1][a.Y].Type == road {
				a.X--
				update = true
			}
		case 3: // right
			if a.X+1 < GridWidth &&
				state.Grid[a.X+1][a.Y].Type == road {
				a.X++
				update = true
			}
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
			state.Actors = append(state.Actors, &Actor{car, int32(x), int32(y)})
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
			tileRect := sdl.Rect{a.X * TileSize, a.Y * TileSize, TileSize, TileSize}
			surface.FillRect(&tileRect, red)
			carRect := sdl.Rect{
				a.X*TileSize + TileSize/4,
				a.Y*TileSize + TileSize/4,
				TileSize - TileSize/2,
				TileSize - TileSize/2,
			}
			surface.FillRect(&carRect, black)

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

	fps := uint32(20)

	running := true
	for running {
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

		for _, a := range state.Actors {
			if TakeAction(a, state) {
				update = true
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
