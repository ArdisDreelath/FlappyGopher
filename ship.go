package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const gravity = 0.2

type ship struct {
	textures []*sdl.Texture
	r        *sdl.Renderer
	time     int
	y, speed float64
}

func newShip(r *sdl.Renderer) (*ship, error) {
	s := ship{textures: make([]*sdl.Texture, 4), r: r, y: 600}
	for i := 1; i <= 4; i++ {
		t, err := img.LoadTexture(r, fmt.Sprintf("res/images/airship%d.png", i))
		if err != nil {
			return nil, fmt.Errorf("Could not load ship image: %v", err)
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

func (s *ship) draw() error {
	s.time++
	s.y -= s.speed
	s.speed += gravity
	if s.y < 50 {
		s.speed = -s.speed * 2 / 3
	}
	rect := &sdl.Rect{W: 50, H: 43, X: 10, Y: int32(768 - s.y)}
	i := s.time / 10 % len(s.textures)
	if err := s.r.Copy(s.textures[i], nil, rect); err != nil {
		return fmt.Errorf("Could not copy background: %v", err)
	}
	return nil
}
