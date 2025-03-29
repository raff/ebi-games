package main

import (
	"bytes"
	"image"

	_ "embed"
	"fmt"
	_ "image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 300
	groundY      = 32 // Adjusted for cartesian coordinates
	gravity      = 0.6
	jumpForce    = 12 // Made positive since y increases upward now
)

var (
	//go:embed assets/ground.png
	groundImage []byte

	//go:embed assets/dino.png
	dinoImage []byte

	//go:embed assets/dino2.png
	dino2Image []byte

	//go:embed assets/dino3.png
	dino3Image []byte

	//go:embed assets/cactus.png
	cactusImage []byte

	//go:embed assets/gameover.png
	gameoverImage []byte
)

type State int

const (
	Idle State = iota
	Playing
	GameOver
)

type Game struct {
	ground    *Sprite
	gameover  *Sprite
	dino      Dino
	obstacles []*Sprite
	score     int
	state     State
}

type Sprite struct {
	x, y   float64
	w, h   float64
	sprite *ebiten.Image
}

type Dino struct {
	sprites []*ebiten.Image
	spi     int

	x, y, sy, vy float64
	jumping      bool
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (g *Game) fix(y float64) float64 {
	return screenHeight - y // Convert cartesian y to screen y
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	if g.state == GameOver {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.reset()
		}
		return nil
	}

	// Update dino
	if ebiten.IsKeyPressed(ebiten.KeySpace) && !g.dino.jumping {
		g.dino.vy = jumpForce
		g.dino.jumping = true
		g.dino.spi = 0
		g.state = Playing
	}

	if g.state == Idle {
		return nil
	}

	g.dino.vy -= gravity // Subtract gravity since y increases upward
	g.dino.y += g.dino.vy

	if g.dino.y < g.dino.sy { // Check against ground in cartesian coordinates
		g.dino.y = g.dino.sy
		g.dino.jumping = false
		g.dino.vy = 0
	}

	if !g.dino.jumping {
		g.dino.spi = 1 + (g.score/4)%2
	}

	// Update obstacles
	for i := range g.obstacles {
		g.obstacles[i].x -= 5
		if g.obstacles[i].x < -50 {
			g.obstacles[i].x = float64(screenWidth + rand.Intn(300))
		}

		// Collision detection (using cartesian coordinates)
		dinoRect := image.Rect(int(g.dino.x), int(g.fix(g.dino.y)), int(g.dino.x)+40, int(g.fix(g.dino.y))+40)
		cactusRect := image.Rect(int(g.obstacles[i].x), int(g.fix(g.obstacles[i].y)), int(g.obstacles[i].x)+30, screenHeight)
		if dinoRect.Overlaps(cactusRect) {
			g.state = GameOver
		}
	}

	g.ground.x -= 5
	if g.ground.x <= -g.ground.w+screenWidth {
		g.ground.x = 0
	}

	g.score++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.state != Idle {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n  Score: %d", g.score))
	}

	if g.state == GameOver {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate((screenWidth-g.gameover.w)/2, (screenHeight-g.gameover.h)/2)
		screen.DrawImage(g.gameover.sprite, op)
		return
	}

	// Draw background
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.ground.x, g.fix(g.ground.y))
	screen.DrawImage(g.ground.sprite, op)

	// Draw obstacles
	for _, cactus := range g.obstacles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(cactus.x, g.fix(cactus.y))
		screen.DrawImage(cactus.sprite, op)
	}

	// Draw dino
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.dino.x, g.fix(g.dino.y))
	screen.DrawImage(g.dino.sprites[g.dino.spi], op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) reset() {
	if g.ground == nil {
		g.ground = loadImage(groundImage)
	}

	g.ground.x = 0
	g.ground.y = g.ground.h

	if g.dino.sprites == nil {
		dino := loadImage(dinoImage)

		g.dino.sprites = append(g.dino.sprites, dino.sprite)
		g.dino.sprites = append(g.dino.sprites, loadImage(dino2Image).sprite)
		g.dino.sprites = append(g.dino.sprites, loadImage(dino3Image).sprite)
		g.dino.sy = dino.h
	}

	g.dino.x = 40
	g.dino.y = g.dino.sy
	g.dino.vy = 0
	g.dino.jumping = false

	if g.gameover == nil {
		g.gameover = loadImage(gameoverImage)
	}

	if g.obstacles == nil {
		g.obstacles = make([]*Sprite, 3)

		for i := range g.obstacles {
			g.obstacles[i] = loadImage(cactusImage)
			g.obstacles[i].x = float64(screenWidth + i*300)
			g.obstacles[i].y = g.obstacles[i].h
		}
	} else {
		for i := range g.obstacles {
			g.obstacles[i].x = float64(screenWidth + i*300)
		}
	}

	g.score = 0
	g.state = Idle
}

func loadImage(imageBytes []byte) *Sprite {
	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(imageBytes))
	if err != nil {
		log.Fatal(err)
	}

	return &Sprite{
		sprite: img,
		w:      float64(img.Bounds().Dx()),
		h:      float64(img.Bounds().Dy()),
	}
}

func main() {
	game := &Game{}
	game.reset()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Chrome Dino Game")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
