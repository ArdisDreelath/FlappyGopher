package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

func run() error {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return fmt.Errorf("Could not initalize SDL: %v", err)
	}
	defer sdl.Quit()

	w, r, err := sdl.CreateWindowAndRenderer(1024, 768, sdl.WINDOW_SHOWN)
	if err != nil {
		return fmt.Errorf("Could not create window: %v", err)
	}
	defer w.Destroy()

	if err := ttf.Init(); err != nil {
		return fmt.Errorf("Could not initialize TTF: %v", err)
	}
	defer ttf.Quit()

	_ = r
	if err := drawTitle(r, "Flappy Gopher"); err != nil {
		return fmt.Errorf("Could not draw a title: %v", err)
	}

	time.Sleep(1 * time.Second)
	s, err := newScene(r)
	if err != nil {
		return fmt.Errorf("Could not create a scene: %v", err)
	}
	defer s.destroy()
	events := make(chan sdl.Event)
	errc := s.run(events)

	runtime.LockOSThread()
	for {
		select {
		case events <- sdl.WaitEvent():
		case err := <-errc:
			return err
		}
	}
}

func drawTitle(r *sdl.Renderer, text string) error {
	r.Clear()
	f, err := ttf.OpenFont("res/fonts/Monoton-Regular.ttf", 615)
	if err != nil {
		return fmt.Errorf("Could not open font: %v", err)
	}
	defer f.Close()

	s, err := f.RenderUTF8Solid(text, sdl.Color{R: 30, G: 30, B: 220, A: 255})
	if err != nil {
		return fmt.Errorf("Could not render a title: %v", err)
	}
	defer s.Free()

	t, err := r.CreateTextureFromSurface(s)
	if err != nil {
		return fmt.Errorf("Could not create texture: %v", err)
	}
	defer t.Destroy()

	if err := r.Copy(t, nil, nil); err != nil {
		return fmt.Errorf("Could not copy a title texture: %v", err)
	}
	r.Present()

	return nil
}
