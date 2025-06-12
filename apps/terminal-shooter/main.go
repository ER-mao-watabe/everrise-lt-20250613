package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/term"
)

const (
	width  = 60
	height = 30
	fps    = 60
)

type Game struct {
	player       *Player
	enemies      []*Enemy
	bullets      []*Bullet
	enemyBullets []*Bullet
	boss         *Boss
	score        int
	stage        int
	gameOver     bool
	stageClear   bool
	frame        int
}

type Player struct {
	x, y   int
	life   int
	symbol rune
}

type Enemy struct {
	x, y      int
	symbol    rune
	moveTimer int
	hp        int
}

type Boss struct {
	x, y        int
	width       int
	height      int
	hp          int
	maxHP       int
	moveDir     int
	attackTimer int
}

type Bullet struct {
	x, y   int
	dx, dy int
	symbol rune
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	game := NewGame()
	game.Run()
}

func NewGame() *Game {
	return &Game{
		player: &Player{
			x:      width / 2,
			y:      height - 3,
			life:   3,
			symbol: 'A',
		},
		enemies:      make([]*Enemy, 0),
		bullets:      make([]*Bullet, 0),
		enemyBullets: make([]*Bullet, 0),
		stage:        1,
	}
}

func (g *Game) Run() {
	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	inputCh := make(chan byte)
	go g.handleInput(inputCh)

	for !g.gameOver {
		select {
		case <-ticker.C:
			g.Update()
			g.Draw()
		case key := <-inputCh:
			g.processInput(key)
		}
	}

	g.showGameOver()
}

func (g *Game) handleInput(ch chan<- byte) {
	buf := make([]byte, 1)
	for {
		os.Stdin.Read(buf)
		ch <- buf[0]
	}
}

func (g *Game) processInput(key byte) {
	switch key {
	case 'q', 'Q':
		g.gameOver = true
	case 'a', 'A':
		if g.player.x > 1 {
			g.player.x--
		}
	case 'd', 'D':
		if g.player.x < width-2 {
			g.player.x++
		}
	case 'w', 'W':
		if g.player.y > 1 {
			g.player.y--
		}
	case 's', 'S':
		if g.player.y < height-2 {
			g.player.y++
		}
	case ' ':
		g.bullets = append(g.bullets, &Bullet{
			x:      g.player.x,
			y:      g.player.y - 1,
			dy:     -1,
			symbol: '|',
		})
	}
}

func (g *Game) Update() {
	g.frame++

	// Update bullets
	for i := len(g.bullets) - 1; i >= 0; i-- {
		b := g.bullets[i]
		b.x += b.dx
		b.y += b.dy
		if b.y < 0 || b.y >= height || b.x < 0 || b.x >= width {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
		}
	}

	// Update enemy bullets
	for i := len(g.enemyBullets) - 1; i >= 0; i-- {
		b := g.enemyBullets[i]
		b.x += b.dx
		b.y += b.dy
		if b.y < 0 || b.y >= height || b.x < 0 || b.x >= width {
			g.enemyBullets = append(g.enemyBullets[:i], g.enemyBullets[i+1:]...)
		}
	}

	// Spawn enemies
	if g.boss == nil && g.frame%60 == 0 && len(g.enemies) < 10 {
		g.enemies = append(g.enemies, &Enemy{
			x:      1 + g.frame%20*2%(width-2),
			y:      1,
			symbol: 'V',
			hp:     1,
		})
	}

	// Update enemies
	for i := len(g.enemies) - 1; i >= 0; i-- {
		e := g.enemies[i]
		e.moveTimer++
		if e.moveTimer > 30 {
			e.moveTimer = 0
			e.y++
			if g.frame%120 == 0 {
				g.enemyBullets = append(g.enemyBullets, &Bullet{
					x:      e.x,
					y:      e.y + 1,
					dy:     1,
					symbol: '.',
				})
			}
		}
		if e.y >= height-1 {
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
		}
	}

	// Check bullet-enemy collisions
	for i := len(g.bullets) - 1; i >= 0; i-- {
		b := g.bullets[i]
		for j := len(g.enemies) - 1; j >= 0; j-- {
			e := g.enemies[j]
			if b.x == e.x && b.y == e.y {
				g.score += 100
				g.enemies = append(g.enemies[:j], g.enemies[j+1:]...)
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				break
			}
		}
	}

	// Check enemy bullet-player collisions
	for i := len(g.enemyBullets) - 1; i >= 0; i-- {
		b := g.enemyBullets[i]
		if b.x == g.player.x && b.y == g.player.y {
			g.player.life--
			g.enemyBullets = append(g.enemyBullets[:i], g.enemyBullets[i+1:]...)
			if g.player.life <= 0 {
				g.gameOver = true
			}
		}
	}

	// Check enemy-player collisions
	for _, e := range g.enemies {
		if e.x == g.player.x && e.y == g.player.y {
			g.player.life--
			if g.player.life <= 0 {
				g.gameOver = true
			}
		}
	}

	// Boss logic
	if g.boss != nil {
		g.updateBoss()
	} else if g.score >= 1000*g.stage && len(g.enemies) == 0 {
		g.spawnBoss()
	}

	// Stage clear
	if g.boss == nil && g.score >= 1000*g.stage && len(g.enemies) == 0 {
		g.stageClear = true
		if g.stageClear && g.frame%120 == 0 {
			g.stage++
			g.stageClear = false
		}
	}
}

func (g *Game) spawnBoss() {
	g.boss = &Boss{
		x:       width/2 - 5,
		y:       3,
		width:   11,
		height:  3,
		hp:      20,
		maxHP:   20,
		moveDir: 1,
	}
}

func (g *Game) updateBoss() {
	// Boss movement
	g.boss.x += g.boss.moveDir
	if g.boss.x <= 1 || g.boss.x+g.boss.width >= width-1 {
		g.boss.moveDir = -g.boss.moveDir
	}

	// Boss attacks
	g.boss.attackTimer++
	if g.boss.attackTimer > 45 {
		g.boss.attackTimer = 0
		for i := 0; i < g.boss.width; i += 2 {
			g.enemyBullets = append(g.enemyBullets, &Bullet{
				x:      g.boss.x + i,
				y:      g.boss.y + g.boss.height,
				dy:     1,
				symbol: 'o',
			})
		}
	}

	// Check bullet-boss collisions
	for i := len(g.bullets) - 1; i >= 0; i-- {
		b := g.bullets[i]
		if b.x >= g.boss.x && b.x < g.boss.x+g.boss.width &&
			b.y >= g.boss.y && b.y < g.boss.y+g.boss.height {
			g.boss.hp--
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
			if g.boss.hp <= 0 {
				g.score += 1000
				g.boss = nil
			}
		}
	}
}

func (g *Game) Draw() {
	clearScreen()
	
	// Draw border
	for i := 0; i < width; i++ {
		fmt.Print("=")
	}
	fmt.Println()

	// Draw game field
	field := make([][]rune, height)
	for i := range field {
		field[i] = make([]rune, width)
		for j := range field[i] {
			if j == 0 || j == width-1 {
				field[i][j] = '|'
			} else {
				field[i][j] = ' '
			}
		}
	}

	// Draw player
	if g.player.y >= 0 && g.player.y < height && g.player.x >= 0 && g.player.x < width {
		field[g.player.y][g.player.x] = g.player.symbol
	}

	// Draw enemies
	for _, e := range g.enemies {
		if e.y >= 0 && e.y < height && e.x >= 0 && e.x < width {
			field[e.y][e.x] = e.symbol
		}
	}

	// Draw boss
	if g.boss != nil {
		for y := 0; y < g.boss.height; y++ {
			for x := 0; x < g.boss.width; x++ {
				if g.boss.y+y >= 0 && g.boss.y+y < height &&
					g.boss.x+x >= 0 && g.boss.x+x < width {
					if y == 0 {
						field[g.boss.y+y][g.boss.x+x] = '['
						if x == g.boss.width-1 {
							field[g.boss.y+y][g.boss.x+x] = ']'
						} else if x > 0 && x < g.boss.width-1 {
							field[g.boss.y+y][g.boss.x+x] = '='
						}
					} else {
						if x == 0 || x == g.boss.width-1 {
							field[g.boss.y+y][g.boss.x+x] = '|'
						} else {
							field[g.boss.y+y][g.boss.x+x] = '#'
						}
					}
				}
			}
		}
	}

	// Draw bullets
	for _, b := range g.bullets {
		if b.y >= 0 && b.y < height && b.x >= 0 && b.x < width {
			field[b.y][b.x] = b.symbol
		}
	}

	// Draw enemy bullets
	for _, b := range g.enemyBullets {
		if b.y >= 0 && b.y < height && b.x >= 0 && b.x < width {
			field[b.y][b.x] = b.symbol
		}
	}

	// Print field
	for _, row := range field {
		fmt.Println(string(row))
	}

	// Draw bottom border and stats
	for i := 0; i < width; i++ {
		fmt.Print("=")
	}
	fmt.Println()
	fmt.Printf("Life: %d | Score: %d | Stage: %d", g.player.life, g.score, g.stage)
	if g.boss != nil {
		fmt.Printf(" | Boss HP: %d/%d", g.boss.hp, g.boss.maxHP)
	}
	if g.stageClear {
		fmt.Print(" | STAGE CLEAR!")
	}
	fmt.Println()
	fmt.Println("Controls: WASD=Move, Space=Shoot, Q=Quit")
}

func (g *Game) showGameOver() {
	clearScreen()
	fmt.Println("=====================================")
	fmt.Println("            GAME OVER                ")
	fmt.Println("=====================================")
	fmt.Printf("        Final Score: %d\n", g.score)
	fmt.Printf("        Stage: %d\n", g.stage)
	fmt.Println("=====================================")
	time.Sleep(3 * time.Second)
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}