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
	ship *ship
}

func newScene(r *sdl.Renderer) (*scene, error) {
	bg, err := img.LoadTexture(r, "res/images/bg1.jpg")
	if err != nil {
		return nil, fmt.Errorf("Could not load background image: %v", err)
	}

	ship, err := newShip(r)
	if err != nil {
		return nil, fmt.Errorf("Could not create ship: %v", err)
	}

	s := scene{r: r, bg: bg, ship: ship}

	return &s, nil
}

func (s *scene) draw() error {
	s.time++
	s.r.Clear()
	if err := s.r.Copy(s.bg, nil, nil); err != nil {
		return fmt.Errorf("Could not copy background: %v", err)
	}

	if err := s.ship.draw(); err != nil {
		return fmt.Errorf("Could not draw ship: %v", err)
	}

	s.r.Present()
	return nil
}

func (s *scene) destroy() {
	s.ship.destroy()
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
