package main

import (
	"context"
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type scene struct {
	time int
	r    *sdl.Renderer
	bg   *sdl.Texture
	ship []*sdl.Texture
}

func newScene(r *sdl.Renderer) (*scene, error) {
	bg, err := img.LoadTexture(r, "res/images/bg1.jpg")
	if err != nil {
		return nil, fmt.Errorf("Could not load background image: %v", err)
	}

	s := scene{r: r, bg: bg, ship: make([]*sdl.Texture, 4)}

	for i := 1; i <= 4; i++ {
		ship, err := img.LoadTexture(r, fmt.Sprintf("res/images/airship%d.png", i))
		if err != nil {
			return nil, fmt.Errorf("Could not load ship image: %v", err)
		}
		s.ship[i-1] = ship
	}

	return &s, nil
}

func (s *scene) draw() error {
	s.time++
	s.r.Clear()
	if err := s.r.Copy(s.bg, nil, nil); err != nil {
		return fmt.Errorf("Could not copy background: %v", err)
	}
	rect := &sdl.Rect{W: 50, H: 43, X: 10, Y: int32(225 + (s.time/5)%100)}
	i := s.time / 10 % len(s.ship)
	if err := s.r.Copy(s.ship[i], nil, rect); err != nil {
		return fmt.Errorf("Could not copy background: %v", err)
	}

	s.r.Present()
	return nil
}

func (s *scene) destroy() {
	for _, ship := range s.ship {
		ship.Destroy()
	}
	s.bg.Destroy()
}

func (s *scene) run(ctx context.Context) <-chan error {
	errc := make(chan error)
	go func() {
		defer close(errc)
		for range time.Tick(10 * time.Millisecond) {
			select {
			case <-ctx.Done():
				return
			default:
				if err := s.draw(); err != nil {
					errc <- err
				}
			}
		}
	}()
	return errc
}
