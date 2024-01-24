package main

import (
	"bytes"
	"flag"
	"image"
	"image/png"
	"io"
	"log"
	"strconv"
	"strings"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//go:embed assets/calc.png
var calc_png []byte

func readImage(r io.Reader) (*ebiten.Image, int, int, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, 0, 0, err
	}

	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	ebimg := ebiten.NewImageFromImage(img)

	return ebimg, iw, ih, nil
}

func coords(x, y, w, h int) image.Rectangle {
	return image.Rect(x, y, x+w, y+h)
}

const None = ebiten.Key(-2)

var buttons = map[ebiten.Key]image.Rectangle{
	ebiten.KeyC:     coords(9, 52, 18, 16),
	ebiten.KeyE:     coords(31, 52, 18, 16),
	ebiten.KeyEnter: coords(53, 52, 18, 16), // =
	ebiten.KeyX:     coords(75, 52, 18, 16), // *

	ebiten.Key7:     coords(9, 74, 18, 16),
	ebiten.Key8:     coords(31, 74, 18, 16),
	ebiten.Key9:     coords(53, 74, 18, 16),
	ebiten.KeySlash: coords(75, 74, 18, 16),

	ebiten.Key4:     coords(9, 96, 18, 16),
	ebiten.Key5:     coords(31, 96, 18, 16),
	ebiten.Key6:     coords(53, 96, 18, 16),
	ebiten.KeyMinus: coords(75, 96, 18, 16),

	ebiten.Key1:     coords(9, 118, 18, 16),
	ebiten.Key2:     coords(31, 118, 18, 16),
	ebiten.Key3:     coords(53, 118, 18, 16),
	ebiten.KeyEqual: coords(75, 118, 18, 38), // +

	ebiten.Key0:      coords(9, 140, 40, 16),
	ebiten.KeyPeriod: coords(53, 140, 18, 16),
}

var displayCoords = coords(13, 33, 77, 9) // display

type Game struct {
	redraw bool
	sel    ebiten.Key

	acc float64
	val string
	op  ebiten.Key

	ima *ebiten.Image
	inv *ebiten.Image
	iw  int
	ih  int
	ww  int
	wh  int

	digits [11]*ebiten.Image

	scale  int
	drawOp ebiten.DrawImageOptions
}

func newGame(scale int) *Game {
	ima, iw, ih, err := readImage(bytes.NewReader(calc_png))
	if err != nil {
		log.Fatal(err)
	}

	inv := ebiten.NewImage(iw, ih)

	var c colorm.ColorM
	c.Scale(-1, -1, -1, 1)
	c.Translate(1, 1, 1, 0)

	colorm.DrawImage(inv, ima, c, &colorm.DrawImageOptions{})

	sw, sh := ebiten.ScreenSizeInFullscreen()

	g := &Game{
		redraw: true,
		op:     None,
		ima:    ima,
		inv:    inv,
		iw:     iw,
		ih:     ih,
		scale:  scale,
		ww:     iw,
		wh:     ih,
	}

	dcoords := func(k ebiten.Key) image.Rectangle {
		min := buttons[k].Min
		return image.Rectangle{Min: image.Point{min.X + 5, min.Y + 5}, Max: image.Point{min.X + 11, min.Y + 13}}
	}

	g.digits[0] = g.ima.SubImage(dcoords(ebiten.Key0)).(*ebiten.Image)
	g.digits[1] = g.ima.SubImage(dcoords(ebiten.Key1)).(*ebiten.Image)
	g.digits[2] = g.ima.SubImage(dcoords(ebiten.Key2)).(*ebiten.Image)
	g.digits[3] = g.ima.SubImage(dcoords(ebiten.Key3)).(*ebiten.Image)
	g.digits[4] = g.ima.SubImage(dcoords(ebiten.Key4)).(*ebiten.Image)
	g.digits[5] = g.ima.SubImage(dcoords(ebiten.Key5)).(*ebiten.Image)
	g.digits[6] = g.ima.SubImage(dcoords(ebiten.Key6)).(*ebiten.Image)
	g.digits[7] = g.ima.SubImage(dcoords(ebiten.Key7)).(*ebiten.Image)
	g.digits[8] = g.ima.SubImage(dcoords(ebiten.Key8)).(*ebiten.Image)
	g.digits[9] = g.ima.SubImage(dcoords(ebiten.Key9)).(*ebiten.Image)
	g.digits[10] = g.ima.SubImage(dcoords(ebiten.KeyPeriod)).(*ebiten.Image)

	if scale != 1 {
		g.drawOp.GeoM.Scale(float64(scale), float64(scale))
		g.ww = g.ww * g.scale
		g.wh = g.wh * g.scale
	}

	if g.ww > sw || g.wh > sh {
		log.Fatal("scale too big")
	}

	return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) drawDisplay(screen *ebiten.Image) {
	var drawOp ebiten.DrawImageOptions
	drawOp.GeoM.Scale(float64(g.scale), float64(g.scale))
	drawOp.GeoM.Translate(float64(displayCoords.Min.X*g.scale), float64(displayCoords.Min.Y*g.scale))

	dim := g.digits[0].Bounds()

	d := g.acc

	if g.val != "" {
		d, _ = strconv.ParseFloat(g.val, 64)
	}

	v := strconv.FormatFloat(d, 'f', -1, 64)
	if strings.HasSuffix(g.val, ".") {
		v += "."
	}

	for _, c := range v {
		if c == '.' {
			c = 10
		} else {
			c -= '0'
		}

		screen.DrawImage(g.digits[c], &drawOp)
		drawOp.GeoM.Translate(float64(dim.Dx()*g.scale), 0)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	screen.DrawImage(g.ima, &g.drawOp)
	g.drawDisplay(screen)

	if v, ok := buttons[g.sel]; ok {
		var drawOp ebiten.DrawImageOptions

		drawOp.GeoM.Scale(float64(g.scale), float64(g.scale))
		drawOp.GeoM.Translate(float64(v.Min.X*g.scale), float64(v.Min.Y*g.scale))

		inv := g.inv.SubImage(v).(*ebiten.Image)
		screen.DrawImage(inv, &drawOp)
	}

	g.redraw = false
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) { // (Q)uit or ESC
		return ebiten.Termination
	}

	p := image.Pt(ebiten.CursorPosition()).Div(g.scale)

	for k, v := range buttons {
		if ebiten.IsKeyPressed(k) || (ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && p.In(v)) {
			if g.sel != k {
				g.sel = k
				g.redraw = true

				g.processKey(k)
			}
			return nil
		}
	}

	if g.sel != None {
		g.sel = None
		g.redraw = true
	}

	return nil
}

func (g *Game) processKey(k ebiten.Key) {
	switch k {
	case ebiten.Key0:
		if g.val != "" {
			g.val += "0"
		}

	case ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9:
		g.val += string(k - ebiten.Key0 + '0')

	case ebiten.KeyPeriod: // .
		if !strings.Contains(g.val, ".") {
			g.val += "."
		}

	case ebiten.KeyE: // exponent

	case ebiten.KeyC: // clear
		g.processOp(k)
		g.op = None

	default:
		if g.op != None {
			g.processOp(g.op)
		} else {
			g.processOp(ebiten.KeyEnter)
		}

		if k == ebiten.KeyEnter {
			g.op = None
		} else {
			g.op = k
		}
	}
}

func (g *Game) processOp(k ebiten.Key) {
	switch k {
	case ebiten.KeyEqual: // +
		v, _ := strconv.ParseFloat(g.val, 64)
		g.acc += v
		g.val = ""

	case ebiten.KeyMinus: // -
		v, _ := strconv.ParseFloat(g.val, 64)
		g.acc -= v
		g.val = ""

	case ebiten.KeyX: // *
		v, _ := strconv.ParseFloat(g.val, 64)
		g.acc *= v
		g.val = ""

	case ebiten.KeySlash: // /
		v, _ := strconv.ParseFloat(g.val, 64)
		g.acc /= v
		g.val = ""

	case ebiten.KeyC: // clear
		if g.val == "" {
			g.acc = 0
		} else {
			g.val = ""
		}

	case ebiten.KeyEnter: // =
		g.acc, _ = strconv.ParseFloat(g.val, 64)
		g.val = ""
	}

	log.Println("acc:", g.acc, "val:", g.val)
}

func main() {
	scale := flag.Int("scale", 4, "Window scale")
	flag.Parse()

	g := newGame(*scale)

	ebiten.SetWindowDecorated(false)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.ww, g.wh)
	ebiten.RunGame(g)
}
