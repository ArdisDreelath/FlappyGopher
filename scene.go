package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/patrikeh/go-deep"

	"github.com/jinzhu/copier"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const nnships = 90

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
	nnships      []*nnship
	creationTime time.Time
}

func newScene(r *sdl.Renderer) (*scene, error) {
	rand.Seed(time.Now().UnixNano())
	bg, err := img.LoadTexture(r, "res/images/bg1.jpg")
	if err != nil {
		return nil, fmt.Errorf("Could not load background image: %v", err)
	}

	ship, err := newShip(r)
	if err != nil {
		return nil, fmt.Errorf("Could not create ship: %v", err)
	}

	s := scene{r: r, bg: bg, ship: ship}

	asteroids, err := newAsteroids(r, ship)
	if err != nil {
		return nil, fmt.Errorf("Could not create asteroids: %v", err)
	}
	s.sceneObjects = append(s.sceneObjects, asteroids)

	files, err := ioutil.ReadDir("gens/best")
	if err != nil {
		return nil, err
	}
	for i, file := range files {
		jsonBytes, _ := ioutil.ReadFile("gens/best/" + file.Name())
		nns, _ := newNNShip(r, i%5+1, asteroids)
		nnDump := deep.Dump{}
		json.Unmarshal(jsonBytes, &nnDump)
		nns.nn = deep.FromDump(&nnDump)
		s.sceneObjects = append(s.sceneObjects, nns)
		s.nnships = append(s.nnships, nns)
		weightScore := WeightScore{Score: 20000, Generation: 0, Weights: nnDump.Weights}
		topWeights.Append(weightScore)
	}

	for i := 0; i < 49; i++ {
		nns, _ := newNNShip(r, i%5+1, asteroids)
		s.sceneObjects = append(s.sceneObjects, nns)
		s.nnships = append(s.nnships, nns)
		nns1 := s.nnships[rand.Intn(len(s.nnships))]
		nns2 := s.nnships[rand.Intn(len(s.nnships))]
		mutateWeights(nns1.nn.Weights(), nns2.nn.Weights(), nns)
	}

	for i := 0; i < nnships; i++ {
		nns, _ := newNNShip(r, i%5+1, asteroids)
		s.sceneObjects = append(s.sceneObjects, nns)
		s.nnships = append(s.nnships, nns)
	}

	s.sceneObjects = append(s.sceneObjects, ship)
	s.creationTime = time.Now()
	os.MkdirAll(fmt.Sprintf("gens/%s", s.creationTime.Format("20060102T150405")), 0770)

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
		ticker := time.Tick(6 * time.Millisecond)
		done := false
		for !done {
			select {
			case e := <-events:
				done = s.handleEvent(e)
			case <-ticker:
				s.update()
				allDead := true
				for _, nns := range s.nnships {
					if !nns.isDead() {
						allDead = false
						break
					}
				}

				if allDead {
					drawTitle(s.r, "GAME OVER")
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
	case *sdl.AudioDeviceEvent:
	case *sdl.MouseButtonEvent:
		if e.Type == sdl.MOUSEBUTTONDOWN {
			s.ship.jump()
		}
	default:
		log.Printf("unknown event %T", event)
	}
	return false
}

var gen = 0
var topWeights = &TopWeights{Limit: 10}

func (s *scene) restart() {
	// get best ai ship
	rand.Seed(time.Now().UnixNano())
	bestScore := -1000
	var nns *nnship
	for _, s := range s.nnships {
		if s.points >= bestScore {
			bestScore = s.points
			nns = s
		}
	}

	gen++
	fmt.Println("Generation : ", gen)
	fmt.Println("Best score: ", bestScore)

	weights := nns.nn.Weights()
	weightScore := WeightScore{Score: bestScore, Generation: gen}
	copier.Copy(&weightScore.Weights, weights)
	topWeights.Append(weightScore)
	bytes, _ := nns.nn.Marshal()
	err := ioutil.WriteFile(fmt.Sprintf("gens/%s/nn_%05d.json", s.creationTime.Format("20060102T150405"), gen), bytes, 0640)
	if err != nil {
		panic(err)
	}

	for _, scor := range topWeights.Scores {
		fmt.Println("Generation: ", scor.Generation, " Points: ", scor.Score)
	}

	for i, s := range s.nnships {
		if i < len(topWeights.Scores) {
			s.nn.ApplyWeights(topWeights.Scores[i].Weights)
		} else {
			if rand.Intn(10) < 3 {
				mutateWeights(topWeights.Scores[rand.Intn(len(topWeights.Scores))].Weights, topWeights.Scores[rand.Intn(len(topWeights.Scores))].Weights, s)
			} else {
				randomizeWeights(topWeights.Scores[rand.Intn(len(topWeights.Scores))].Weights, rand.Intn(10)+1, s)
			}
		}
	}

	for _, o := range s.sceneObjects {
		o.reset()
	}
}

func randomizeWeights(weights [][][]float64, chance int, nns *nnship) {
	shipWeigths := nns.nn.Weights()
	for x := range weights {
		for y := range weights[x] {
			for z := range weights[x][y] {
				if rand.Intn(chance) == 0 {
					shipWeigths[x][y][z] = weights[x][y][z] * ((rand.Float64()) + 0.5)
				}
			}
		}
	}
	nns.nn.ApplyWeights(shipWeigths)
}

func reverseWeights(weights [][][]float64, chance int) {
	for x := range weights {
		for y := range weights[x] {
			for z := range weights[x][y] {
				if rand.Intn(chance) == 0 {
					weights[x][y][z] = -weights[x][y][z]
				}
			}
		}
	}
}

func mutateWeights(weights1, weights2 [][][]float64, nnsout *nnship) {
	weightsout := nnsout.nn.Weights()
	for x := range weights1 {
		for y := range weights1[x] {
			for z := range weights1[x][y] {
				if rand.Intn(2) == 0 {
					weightsout[x][y][z] = weights1[x][y][z]
				} else {
					weightsout[x][y][z] = weights2[x][y][z]
				}
			}
		}
	}
	nnsout.nn.ApplyWeights(weightsout)
}
