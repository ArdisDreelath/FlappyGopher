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

	r          *sdl.Renderer
	speed      float64
	x, y, w, h int32
	dead       bool
}

func newShip(r *sdl.Renderer) (*ship, error) {
	s := ship{textures: make([]*sdl.Texture, 4), r: r, y: 600, w: 50, h: 43, x: 15}
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

func (s *ship) update() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.time++
	s.y -= int32(s.speed)
	s.speed += gravity
	if s.y > 768 {
		s.dead = true
		return nil
	}
	if s.y < s.h {
		s.dead = true
		return nil
	}

	return nil
}

func (s *ship) draw() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rect := &sdl.Rect{W: s.w, H: s.h, X: s.x, Y: int32(768 - s.y)}
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

func (s *ship) reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.y = 600
	s.dead = false
	s.speed = 0
	return nil
}

func (s *ship) hit() {
	s.dead = true
}
