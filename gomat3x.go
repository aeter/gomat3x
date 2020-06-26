package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	wg       sync.WaitGroup
	alphanum = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type position struct {
	x int
	y int
}

type snake struct {
	head   position
	length int
}

func (s *snake) move() {
	s.head.y++
}

func (s *snake) draw() {
	defer wg.Done()
	termbox.SetCell(s.head.x, s.head.y, randomChar(), termbox.ColorWhite, termbox.ColorBlack)
	for i := 1; i < s.length; i++ {
		termbox.SetCell(s.head.x, s.head.y-i, randomChar(), termbox.ColorGreen, termbox.ColorBlack)
	}
}

func (s *snake) outOfScreen() bool {
	_, h := termbox.Size()
	return s.head.y-s.length > h
}

func randomChar() rune {
	return alphanum[rand.Intn(len(alphanum))]
}

func addSnakes(snakes *[]snake) {
	// Find which snakes intersect the first, top of screen row. These
	// positions are forbidden for spawning new snakes.
	forbiddenSpawnX := make(map[int]bool)
	for _, snake := range *snakes {
		if snake.head.y-snake.length <= 0 {
			forbiddenSpawnX[snake.head.x] = true
		}
	}

	w, _ := termbox.Size()
	for i := 0; i < w; i += 2 {
		if rand.Intn(100) < 8 { // N percent chance of a new snake
			if !forbiddenSpawnX[i] {
				newSnake := snake{head: position{x: i, y: 0}, length: rand.Intn(10) + 3}
				*snakes = append(*snakes, newSnake)
			}
		}
	}
}

func removeOutOfScreenSnakes(in []snake) []snake {
	out := make([]snake, 0)
	for _, s := range in {
		if !s.outOfScreen() {
			out = append(out, s)
		}
	}
	return out
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	snakes := []snake{snake{head: position{x: 0, y: 1}, length: 3}}

loop:
	for {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeyCtrlC {
				break loop // bye bye
			}
			if ev.Type == termbox.EventResize {
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				termbox.Flush()
			}
			if ev.Type == termbox.EventError {
				panic(ev.Err)
			}
		default:
			time.Sleep(100 * time.Millisecond)
			termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)
			addSnakes(&snakes)
			snakes = removeOutOfScreenSnakes(snakes)
			for i := range snakes {
				snakes[i].move()
			}
			for i := range snakes {
				wg.Add(1)
				go snakes[i].draw() // drawing is expensive so using goroutines
			}
			wg.Wait()
			termbox.Flush()
		}
	}
}
