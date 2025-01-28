package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	width      = 25
	height     = 20
	foodChar   = '◆'
	snakeChar  = '▣'
	borderChar = '■'
)

type Position struct {
	X int
	Y int
}

var (
	snake     []Position
	food      Position
	direction string
	gameOver  bool
	score     int
	inputCh   = make(chan string)
	quitCh    = make(chan struct{})
)

// Color scheme
const (
	ColorSnake    = termbox.ColorGreen
	ColorFood     = termbox.ColorRed
	ColorBorder   = termbox.ColorCyan
	ColorScore    = termbox.ColorWhite | termbox.AttrBold
	ColorGameOver = termbox.ColorRed | termbox.AttrBold
	ColorText     = termbox.ColorYellow
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	showWelcomeScreen()
	initializeGame()
	go readInput()

	gameLoop()
}

func showWelcomeScreen() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawCenteredText("SNAKE GAME", 5, ColorText)
	drawCenteredText("Use arrow keys to move", 8, ColorText)
	drawCenteredText("Collect the "+string(foodChar)+" to grow", 9, ColorFood)
	drawCenteredText("Avoid walls and yourself!", 10, ColorText)
	drawCenteredText("Press any key to start", 13, ColorText)
	termbox.Flush()

	// Wait for any key
	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			return
		}
	}
}

func initializeGame() {
	rand.Seed(time.Now().UnixNano())
	snake = []Position{{X: width/2 - 2, Y: height / 2}}
	direction = "RIGHT"
	gameOver = false
	score = 0
	placeFood()
}

func gameLoop() {
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !gameOver {
				move()
				checkCollisions()
				draw()
			}
		case key := <-inputCh:
			handleInput(key)
		case <-quitCh:
			return
		}
	}
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawBorder()
	drawSnake()
	drawFood()
	drawScore()

	if gameOver {
		drawGameOver()
	}

	termbox.Flush()
}

func drawBorder() {
	// Top and bottom borders
	for x := 0; x < width+2; x++ {
		termbox.SetCell(x, 0, borderChar, ColorBorder, termbox.ColorDefault)
		termbox.SetCell(x, height+1, borderChar, ColorBorder, termbox.ColorDefault)
	}

	// Side borders
	for y := 1; y <= height; y++ {
		termbox.SetCell(0, y, borderChar, ColorBorder, termbox.ColorDefault)
		termbox.SetCell(width+1, y, borderChar, ColorBorder, termbox.ColorDefault)
	}
}

func drawSnake() {
	for i, pos := range snake {
		color := ColorSnake
		if i == 0 {
			color = termbox.ColorGreen | termbox.AttrBold
		}
		termbox.SetCell(pos.X+1, pos.Y+1, snakeChar, color, termbox.ColorDefault)
	}
}

func drawFood() {
	termbox.SetCell(food.X+1, food.Y+1, foodChar, ColorFood, termbox.ColorDefault)
}

func drawScore() {
	scoreText := fmt.Sprintf(" SCORE: %d ", score)
	instructions := " PRESS Q TO QUIT "

	// Draw score box
	for i, ch := range scoreText {
		termbox.SetCell(i+2, height+3, ch, ColorScore, termbox.ColorBlue)
	}

	// Draw instructions
	for i, ch := range instructions {
		termbox.SetCell(width+2-len(instructions)+i, height+3, ch, ColorText, termbox.ColorBlue)
	}
}

func drawGameOver() {
	message := []string{
		"╔═════════════════════╗",
		"║      GAME OVER      ║",
		"║                     ║",
		"║   FINAL SCORE: %3d  ║",
		"║                     ║",
		"║  PRESS Q TO QUIT    ║",
		"╚═════════════════════╝",
	}

	yStart := (height - len(message)) / 2
	for i, line := range message {
		if i == 3 {
			line = fmt.Sprintf(line, score)
		}
		x := (width-len(line))/2 + 1
		for j, ch := range line {
			termbox.SetCell(x+j, yStart+i, ch, ColorGameOver, termbox.ColorDefault)
		}
	}
}

func drawCenteredText(text string, y int, color termbox.Attribute) {
	x := (width-len(text))/2 + 1
	for i, ch := range text {
		termbox.SetCell(x+i, y, ch, color, termbox.ColorDefault)
	}
}

func readInput() {
	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch {
			case ev.Key == termbox.KeyArrowUp:
				inputCh <- "UP"
			case ev.Key == termbox.KeyArrowDown:
				inputCh <- "DOWN"
			case ev.Key == termbox.KeyArrowLeft:
				inputCh <- "LEFT"
			case ev.Key == termbox.KeyArrowRight:
				inputCh <- "RIGHT"
			case ev.Ch == 'q' || ev.Key == termbox.KeyEsc:
				close(quitCh)
				return
			}
		}
	}
}

func handleInput(key string) {
	switch key {
	case "UP":
		if direction != "DOWN" {
			direction = key
		}
	case "DOWN":
		if direction != "UP" {
			direction = key
		}
	case "LEFT":
		if direction != "RIGHT" {
			direction = key
		}
	case "RIGHT":
		if direction != "LEFT" {
			direction = key
		}
	}
}

func move() {
	head := snake[0]
	newHead := Position{X: head.X, Y: head.Y}

	switch direction {
	case "UP":
		newHead.Y--
	case "DOWN":
		newHead.Y++
	case "LEFT":
		newHead.X--
	case "RIGHT":
		newHead.X++
	}

	snake = append([]Position{newHead}, snake...)

	if newHead.X == food.X && newHead.Y == food.Y {
		score += 10
		placeFood()
	} else {
		snake = snake[:len(snake)-1]
	}
}

func checkCollisions() {
	head := snake[0]

	// Wall collision
	if head.X < 0 || head.X >= width || head.Y < 0 || head.Y >= height {
		gameOver = true
	}

	// Self collision
	for _, segment := range snake[1:] {
		if head.X == segment.X && head.Y == segment.Y {
			gameOver = true
		}
	}
}

func placeFood() {
	for {
		food = Position{
			X: rand.Intn(width),
			Y: rand.Intn(height),
		}

		// Check if food spawned on snake
		validPosition := true
		for _, segment := range snake {
			if food.X == segment.X && food.Y == segment.Y {
				validPosition = false
				break
			}
		}

		if validPosition {
			break
		}
	}
}
