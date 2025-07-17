package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	// "path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nfnt/resize"
)

type Game struct {
	canvas *ebiten.Image
	redraw bool

	filename string

	iw int
	ih int

	ww int
	wh int

	drawOp ebiten.DrawImageOptions
}

func NewGame() *Game {
	return &Game{}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()
	if x < 0 || y < 0 || x >= g.ww || y >= g.wh {
		return nil
	}

	ebiten.SetWindowTitle(fmt.Sprintf("%s: w=%d h=%d x=%d y=%d", g.filename, g.ww, g.wh, x, y))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	screen.DrawImage(g.canvas, &g.drawOp)
}

func getFile(fpath string) (*os.File, int64, error) {

	f, err := os.Open(fpath)
	if err != nil {
		return nil, 0, err
	}

	finf, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}

	return f, finf.Size(), nil
}

func (g *Game) ReadImage(imafile string, w, h int) error {
	f, _, err := getFile(imafile)
	if err != nil {
		return err
	}

	defer f.Close()

	ima, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	iw, ih := ima.Bounds().Dx(), ima.Bounds().Dy()

	if iw > w || ih > h {
		// Resize the image if it exceeds the maximum dimensions
		if iw > w {
			ih = ih * w / iw
			iw = w
		}
		if ih > h {
			iw = iw * h / ih
			ih = h
		}

		ima = resize.Resize(uint(iw), uint(ih), ima, resize.Lanczos3)
	}

	g.canvas = ebiten.NewImageFromImage(ima)
	g.iw, g.ih = iw, ih
	g.ww, g.wh = iw, ih
	g.filename = imafile
	g.redraw = true
	return nil
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tima [options] <image-file>...")
		os.Exit(1)
	}

	filename := flag.Arg(0)

	g := NewGame()

	w, h := ebiten.ScreenSizeInFullscreen()
	if err := g.ReadImage(filename, w, h); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading image: %v\n", err)
		os.Exit(1)
	}

	ebiten.SetWindowTitle(g.filename)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.ww, g.wh)
	ebiten.RunGame(g)
}
