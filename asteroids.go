package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Tarliton/collision2d"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type asteroids struct {
	mu       sync.RWMutex
	time     int
	textures []*sdl.Texture
	r        *sdl.Renderer
	list     []*asteroid
	ship     *ship
}

type asteroid struct {
	x, y     int32
	speed, r float64
	t        *sdl.Texture
}

func newAsteroids(r *sdl.Renderer, s *ship) (*asteroids, error) {
	a := asteroids{r: r, ship: s, textures: make([]*sdl.Texture, 1), list: make([]*asteroid, 0, 30)}
	t, err := img.LoadTexture(r, "res/images/asteroid1.png")
	if err != nil {
		return nil, fmt.Errorf("Could not load asteroid texture: %v", err)
	}
	a.textures[0] = t
	go func() {
		for {
			time.Sleep(time.Millisecond * 750)
			//size := 1 + rand.Intn(5)
			size := 3
			r := float64(20 + 10*size)
			//speed := float64(10 - size)
			speed := float64(9)
			a.createAsteroid(int32(rand.Intn(786)), r, speed, a.textures[0])
		}
	}()
	return &a, nil
}

func (a *asteroids) update() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.time++
	tmpList := make([]*asteroid, 0, len(a.list))

	for _, ast := range a.list {
		ast.x -= int32(ast.speed)
		if ast.x+2*int32(ast.r) > 0 {
			tmpList = append(tmpList, ast)
		}
	}

	a.list = tmpList
	a.checkCollisions()

	return nil
}

func (a *asteroids) draw() error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, ast := range a.list {
		rect := &sdl.Rect{W: 2 * int32(ast.r), H: 2 * int32(ast.r), X: ast.x - int32(ast.r), Y: int32(768 - ast.y - int32(ast.r))}
		if err := a.r.Copy(ast.t, nil, rect); err != nil {
			return fmt.Errorf("Could not copy asteroid texture: %v", err)
		}
	}
	return nil
}

func (a *asteroids) reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.time = 0
	a.list = make([]*asteroid, 0, 30)
	return nil
}

func (a *asteroids) createAsteroid(y int32, r float64, speed float64, texture *sdl.Texture) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.list = append(a.list, &asteroid{t: texture, x: 2000, y: y, r: r, speed: speed})
}

func (a *asteroids) checkCollisions() {
	a.ship.mu.RLock()
	defer a.ship.mu.RUnlock()

	wg := sync.WaitGroup{}
	wg.Add(len(a.list))
	shipBox := collision2d.NewBox(collision2d.NewVector(float64(a.ship.x-a.ship.w/2), float64(a.ship.y-a.ship.h/2)), float64(a.ship.w), float64(a.ship.h))
	hit := false

	for _, ast := range a.list {
		go func(ast *asteroid) {
			astCircle := collision2d.NewCircle(collision2d.NewVector(float64(ast.x), float64(ast.y)), ast.r)
			inCollision, _ := collision2d.TestPolygonCircle(shipBox.ToPolygon(), astCircle)
			if inCollision {
				hit = true
			}
			wg.Done()
		}(ast)
	}
	wg.Wait()
	if hit {
		a.ship.hit()
	}
}

func (a *asteroids) checkCollisionWithPoint(x, y int) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, ast := range a.list {
		astCircle := collision2d.NewCircle(collision2d.NewVector(float64(ast.x), float64(ast.y)), ast.r)
		inCollision := collision2d.PointInCircle(collision2d.NewVector(float64(x), float64(y)), astCircle)
		if inCollision {
			return true
		}
	}

	return false
}

func (a *asteroids) checkCollisionsWithRect(x, y, w, h int32) bool {
	shipBox := collision2d.NewBox(collision2d.NewVector(float64(x-w/2), float64(y-h/2)), float64(w), float64(h))

	for _, ast := range a.list {
		astCircle := collision2d.NewCircle(collision2d.NewVector(float64(ast.x), float64(ast.y)), ast.r)
		inCollision, _ := collision2d.TestPolygonCircle(shipBox.ToPolygon(), astCircle)
		if inCollision {
			return true
		}
	}

	return false
}
