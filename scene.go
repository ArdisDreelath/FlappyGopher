package main

import (
	"fmt"
	"log"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type sceneObject interface {
	update() error
	draw() error
	reset() error
}

type scene struct {
	time         int
	r            *sdl.Renderer
	bg           *sdl.Texture
	ship         *ship
	sceneObjects []sceneObject
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
	s.sceneObjects = append(s.sceneObjects, ship)

	asteroids, err := newAsteroids(r, ship)
	if err != nil {
		return nil, fmt.Errorf("Could not create asteroids: %v", err)
	}
	s.sceneObjects = append(s.sceneObjects, asteroids)

	return &s, nil
}

func (s *scene) update() {
	for _, o := range s.sceneObjects {
		o.update()
	}
}

func (s *scene) draw() error {
	s.time++
	s.r.Clear()
	if err := s.r.Copy(s.bg, nil, nil); err != nil {
		return fmt.Errorf("Could not copy background: %v", err)
	}

	for _, o := range s.sceneObjects {
		o.draw()
	}

	s.r.Present()
	return nil
}

func (s *scene) destroy() {
	s.ship.destroy()
	s.bg.Destroy()
}

func (s *scene) run(events <-chan sdl.Event) <-chan error {
	errc := make(chan error)

	go func() {
		defer close(errc)
		ticker := time.Tick(10 * time.Millisecond)
		done := false
		for !done {
			select {
			case e := <-events:
				done = s.handleEvent(e)
			case <-ticker:
				s.update()
				if s.ship.isDead() {
					drawTitle(s.r, "GAME OVER")
					time.Sleep(time.Second)
					s.restart()
				}

				if err := s.draw(); err != nil {
					errc <- err
				}
			}
		}
	}()
	return errc
}

func (s *scene) handleEvent(event sdl.Event) bool {
	switch e := event.(type) {
	case *sdl.QuitEvent:
		return true
	case *sdl.MouseMotionEvent:
	case *sdl.WindowEvent:
	case *sdl.MouseButtonEvent:
		if e.Type == sdl.MOUSEBUTTONDOWN {
			s.ship.jump()
		}
	default:
		log.Printf("unknown event %T", event)
	}
	return false
}

func (s *scene) restart() {
	for _, o := range s.sceneObjects {
		o.reset()
	}
}
