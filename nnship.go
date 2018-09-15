package main

import (
	"fmt"
	"sync"

	"github.com/patrikeh/go-deep"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

//const gravity = 0.1

var nnconfig = deep.Config{
	/* Input dimensionality */
	Inputs: 8,
	/* Two hidden layers consisting of two neurons each, and a single output */
	Layout: []int{4, 4, 1},
	/* Activation functions: Sigmoid, Tanh, ReLU, Linear */
	Activation: deep.ActivationSigmoid,
	/* Determines output layer activation & loss function:
	ModeRegression: linear outputs with MSE loss
	ModeMultiClass: softmax output with Cross Entropy loss
	ModeMultiLabel: sigmoid output with Cross Entropy loss
	ModeBinary: sigmoid output with binary CE loss */
	Mode: deep.ModeBinary,
	/* Weight initializers: {deep.NewNormal(μ, σ), deep.NewUniform(μ, σ)} */
	Weight: deep.NewNormal(1.0, 0.0),
	/* Apply bias */
	Bias: true,
}

type nnship struct {
	mu sync.RWMutex

	time    int
	texture *sdl.Texture

	r          *sdl.Renderer
	speed      float64
	x, y, w, h int32
	dead       bool
	points     int
	asteroids  *asteroids
	nn         *deep.Neural
}

var loadedTextures []*sdl.Texture

func newNNShip(r *sdl.Renderer, texture int, asteroids *asteroids) (*nnship, error) {
	s := nnship{r: r, y: 600, w: 50, h: 43, x: 15, asteroids: asteroids}
	if len(loadedTextures) == 0 {
		for i := 1; i <= 5; i++ {
			t, err := img.LoadTexture(r, fmt.Sprintf("res/images/nnShip%d.png", i))
			if err != nil {
				return nil, fmt.Errorf("Could not load nnship texture: %d, %v", i, err)
			}
			loadedTextures = append(loadedTextures, t)
		}
	}
	s.texture = loadedTextures[texture-1]
	s.nn = deep.NewNeural(&nnconfig)
	return &s, nil
}

func (s *nnship) destroy() {
	s.texture.Destroy()
}

func (s *nnship) update() error {
	if s.dead {
		return nil
	}

	if s.asteroids.checkCollisionsWithRect(s.x, s.y, s.w, s.h) {
		s.dead = true
		return nil
	}

	s.points++

	if s.shouldJump() {
		s.jump()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.time++
	s.y -= int32(s.speed)
	s.speed += gravity
	if s.y > 768 {
		s.points -= 300
		s.dead = true
		return nil
	}
	if s.y < s.h {
		s.points -= 300
		s.dead = true
		return nil
	}

	return nil
}

func (s *nnship) draw() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.dead {
		return nil
	}
	rect := &sdl.Rect{W: s.w, H: s.h, X: s.x, Y: int32(768 - s.y)}
	if err := s.r.Copy(s.texture, nil, rect); err != nil {
		return fmt.Errorf("Could not ship texture: %v", err)
	}
	return nil
}

func (s *nnship) jump() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speed = -5
}

func (s *nnship) isDead() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dead
}

func (s *nnship) reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.y = 600
	s.dead = false
	s.speed = 0
	s.points = 0
	return nil
}

func (s *nnship) hit() {
	s.dead = true
}

func (s *nnship) radar(angle float32) int {
	for i := 3; i < 901; i += 3 {
		y := int(s.y) + int(angle*float32(i))
		x := int(s.x) + i
		if y < 0 {
			return 9001
		}

		if y > 768 {
			return 9001
		}

		if s.asteroids.checkCollisionWithPoint(x, y) {
			return i
		}
	}

	return 9001
}

func (s *nnship) shouldJump() bool {
	c := make(chan bool)
	go func(s *nnship) {
		//inputs := []float64{float64(s.y), float64(s.speed), float64(1), float64(1), float64(1), float64(1)}
		inputs := []float64{
			float64(s.y),
			float64(s.speed),
			float64(s.radar(2)),
			float64(s.radar(1)),
			float64(s.radar(0.5)),
			float64(s.radar(-0.5)),
			float64(s.radar(-1)),
			float64(s.radar(-2)),
		}
		res := s.nn.Predict(inputs)
		c <- (res[0] > 0.5)
	}(s)
	return <-c
}
