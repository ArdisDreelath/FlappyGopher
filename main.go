package main

import (
	"fmt"
	"os"
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
	if err := drawTitle(r); err != nil {
		return fmt.Errorf("Could not draw a title: %v", err)
	}
	time.Sleep(5 * time.Second)

	return nil
}

func drawTitle(r *sdl.Renderer) error {
	r.Clear()
	f, err := ttf.OpenFont("res/fonts/Monoton-Regular.ttf", 315)
	if err != nil {
		return fmt.Errorf("Could not open font: %v", err)
	}
	defer f.Close()

	s, err := f.RenderUTF8Solid("Flappy Gopher", sdl.Color{R: 30, G: 30, B: 220, A: 255})
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
