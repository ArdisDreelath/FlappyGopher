package main

import (
	"fmt"
	"sync"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const gravity = 0.1

type ship struct {
	mu sync.RWMutex

	time     int
	textures []*sdl.Texture

	r        *sdl.Renderer
	y, speed float64
	dead     bool
}

func newShip(r *sdl.Renderer) (*ship, error) {
	s := ship{textures: make([]*sdl.Texture, 4), r: r, y: 600}
	for i := 1; i <= 4; i++ {
		t, err := img.LoadTexture(r, fmt.Sprintf("res/images/airship%d.png", i))
		if err != nil {
			return nil, fmt.Errorf("Could not load ship texture: %v", err)
		}
		s.textures[i-1] = t
	}
	return &s, nil
}

func (s *ship) destroy() {
	for _, t := range s.textures {
		t.Destroy()
	}
}

func (s *ship) update() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.time++
	s.y -= s.speed
	s.speed += gravity
	if s.y < 30 {
		s.y = 600
		s.dead = true
	}
}

func (s *ship) draw() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rect := &sdl.Rect{W: 50, H: 43, X: 10, Y: int32(768 - s.y)}
	i := s.time / 10 % len(s.textures)
	if err := s.r.Copy(s.textures[i], nil, rect); err != nil {
		return fmt.Errorf("Could not ship texture: %v", err)
	}
	return nil
}

func (s *ship) jump() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speed = -5
}

func (s *ship) isDead() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dead
}

func (s *ship) restart() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.y = 600
	s.dead = false
	s.speed = 0
}
