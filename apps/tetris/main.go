package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	boardWidth  = 10
	boardHeight = 20
	blockSize   = 2
)

type Point struct {
	X, Y int
}

type Tetromino struct {
	Blocks []Point
	Color  tcell.Color
}

type Game struct {
	screen       tcell.Screen
	board        [boardHeight][boardWidth]tcell.Color
	currentPiece *Tetromino
	currentX     int
	currentY     int
	score        int
	level        int
	lines        int
	gameOver     bool
	fallTimer    time.Time
	fallSpeed    time.Duration
}

var tetrominos = [][]Point{
	{{0, 0}, {1, 0}, {0, 1}, {1, 1}}, // O
	{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, // I
	{{0, 0}, {1, 0}, {2, 0}, {1, 1}}, // T
	{{0, 0}, {1, 0}, {1, 1}, {2, 1}}, // S
	{{1, 0}, {2, 0}, {0, 1}, {1, 1}}, // Z
	{{0, 0}, {0, 1}, {1, 1}, {2, 1}}, // J
	{{2, 0}, {0, 1}, {1, 1}, {2, 1}}, // L
}

var colors = []tcell.Color{
	tcell.ColorYellow,
	tcell.ColorAqua,
	tcell.ColorPurple,
	tcell.ColorGreen,
	tcell.ColorRed,
	tcell.ColorBlue,
	tcell.ColorOlive,
}

func NewGame() *Game {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	screen.Clear()

	return &Game{
		screen:    screen,
		score:     0,
		level:     1,
		lines:     0,
		fallSpeed: time.Millisecond * 1000,
		fallTimer: time.Now(),
	}
}

func (g *Game) spawnPiece() {
	idx := rand.Intn(len(tetrominos))
	g.currentPiece = &Tetromino{
		Blocks: make([]Point, len(tetrominos[idx])),
		Color:  colors[idx],
	}
	copy(g.currentPiece.Blocks, tetrominos[idx])
	g.currentX = boardWidth/2 - 1
	g.currentY = 0

	if !g.isValidPosition(g.currentX, g.currentY, g.currentPiece.Blocks) {
		g.gameOver = true
	}
}

func (g *Game) rotatePiece() {
	if g.currentPiece == nil {
		return
	}

	rotated := make([]Point, len(g.currentPiece.Blocks))
	centerX, centerY := 1.5, 1.5

	for i, block := range g.currentPiece.Blocks {
		x := float64(block.X) - centerX
		y := float64(block.Y) - centerY
		rotated[i].X = int(-y + centerX)
		rotated[i].Y = int(x + centerY)
	}

	if g.isValidPosition(g.currentX, g.currentY, rotated) {
		g.currentPiece.Blocks = rotated
	}
}

func (g *Game) isValidPosition(x, y int, blocks []Point) bool {
	for _, block := range blocks {
		newX := x + block.X
		newY := y + block.Y

		if newX < 0 || newX >= boardWidth || newY >= boardHeight {
			return false
		}

		if newY >= 0 && g.board[newY][newX] != tcell.ColorDefault {
			return false
		}
	}
	return true
}

func (g *Game) lockPiece() {
	for _, block := range g.currentPiece.Blocks {
		x := g.currentX + block.X
		y := g.currentY + block.Y
		if y >= 0 && y < boardHeight && x >= 0 && x < boardWidth {
			g.board[y][x] = g.currentPiece.Color
		}
	}
	g.clearLines()
	g.spawnPiece()
}

func (g *Game) clearLines() {
	linesCleared := 0
	for y := boardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < boardWidth; x++ {
			if g.board[y][x] == tcell.ColorDefault {
				full = false
				break
			}
		}

		if full {
			for moveY := y; moveY > 0; moveY-- {
				for x := 0; x < boardWidth; x++ {
					g.board[moveY][x] = g.board[moveY-1][x]
				}
			}
			for x := 0; x < boardWidth; x++ {
				g.board[0][x] = tcell.ColorDefault
			}
			linesCleared++
			y++
		}
	}

	if linesCleared > 0 {
		g.lines += linesCleared
		g.score += linesCleared * 100 * g.level

		if g.lines >= g.level*10 {
			g.level++
			g.fallSpeed = time.Millisecond * time.Duration(1000-g.level*50)
			if g.fallSpeed < 100*time.Millisecond {
				g.fallSpeed = 100 * time.Millisecond
			}
		}
	}
}

func (g *Game) update() {
	if g.gameOver {
		return
	}

	if time.Since(g.fallTimer) >= g.fallSpeed {
		if g.isValidPosition(g.currentX, g.currentY+1, g.currentPiece.Blocks) {
			g.currentY++
		} else {
			g.lockPiece()
		}
		g.fallTimer = time.Now()
	}
}

func (g *Game) draw() {
	g.screen.Clear()

	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			color := g.board[y][x]
			if color != tcell.ColorDefault {
				g.drawBlock(x*blockSize+1, y+1, color)
			}
		}
	}

	if g.currentPiece != nil && !g.gameOver {
		for _, block := range g.currentPiece.Blocks {
			x := (g.currentX + block.X) * blockSize + 1
			y := g.currentY + block.Y + 1
			g.drawBlock(x, y, g.currentPiece.Color)
		}
	}

	for y := 0; y <= boardHeight; y++ {
		g.screen.SetContent(0, y, '│', nil, tcell.StyleDefault)
		g.screen.SetContent(boardWidth*blockSize+1, y, '│', nil, tcell.StyleDefault)
	}
	for x := 0; x <= boardWidth*blockSize+1; x++ {
		g.screen.SetContent(x, boardHeight+1, '─', nil, tcell.StyleDefault)
	}

	infoX := boardWidth*blockSize + 4
	g.drawText(infoX, 2, fmt.Sprintf("Score: %d", g.score))
	g.drawText(infoX, 4, fmt.Sprintf("Level: %d", g.level))
	g.drawText(infoX, 6, fmt.Sprintf("Lines: %d", g.lines))

	if g.gameOver {
		g.drawText(boardWidth/2-4, boardHeight/2, "GAME OVER")
		g.drawText(boardWidth/2-6, boardHeight/2+2, "Press Q to quit")
	}

	g.screen.Show()
}

func (g *Game) drawBlock(x, y int, color tcell.Color) {
	style := tcell.StyleDefault.Background(color)
	g.screen.SetContent(x, y, ' ', nil, style)
	g.screen.SetContent(x+1, y, ' ', nil, style)
}

func (g *Game) drawText(x, y int, text string) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for i, ch := range text {
		g.screen.SetContent(x+i, y, ch, nil, style)
	}
}

func (g *Game) handleInput() {
	for {
		ev := g.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if g.gameOver && ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
				g.saveHighScore()
				g.screen.Fini()
				os.Exit(0)
			}

			if !g.gameOver && g.currentPiece != nil {
				switch ev.Key() {
				case tcell.KeyLeft:
					if g.isValidPosition(g.currentX-1, g.currentY, g.currentPiece.Blocks) {
						g.currentX--
					}
				case tcell.KeyRight:
					if g.isValidPosition(g.currentX+1, g.currentY, g.currentPiece.Blocks) {
						g.currentX++
					}
				case tcell.KeyDown:
					if g.isValidPosition(g.currentX, g.currentY+1, g.currentPiece.Blocks) {
						g.currentY++
						g.score++
					}
				case tcell.KeyUp:
					g.rotatePiece()
				case tcell.KeyRune:
					if ev.Rune() == ' ' {
						for g.isValidPosition(g.currentX, g.currentY+1, g.currentPiece.Blocks) {
							g.currentY++
							g.score += 2
						}
					}
				case tcell.KeyEscape, tcell.KeyCtrlC:
					g.saveHighScore()
					g.screen.Fini()
					os.Exit(0)
				}
			}
		case *tcell.EventResize:
			g.screen.Sync()
		}
	}
}

func (g *Game) saveHighScore() {
	highScore := g.loadHighScore()
	if g.score > highScore {
		file, err := os.Create("highscore.txt")
		if err == nil {
			defer file.Close()
			fmt.Fprintf(file, "%d", g.score)
		}
	}
}

func (g *Game) loadHighScore() int {
	data, err := os.ReadFile("highscore.txt")
	if err != nil {
		return 0
	}
	var score int
	fmt.Sscanf(string(data), "%d", &score)
	return score
}

func main() {
	rand.Seed(time.Now().UnixNano())
	game := NewGame()
	defer game.screen.Fini()

	game.spawnPiece()
	go game.handleInput()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		game.update()
		game.draw()
	}
}