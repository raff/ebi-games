package util

import (
	"errors"
	"image"
	"image/png"
	"io"

	"github.com/hajimehoshi/ebiten/v2"
)

// struct Tiles contains a list of image tiles.
// It also contains the tile dimension
type Tiles struct {
	Width   int             // tile width
	Height  int             // tile height
	Columns int             // number of tiles per row
	Rows    int             // number of rows
	List    []*ebiten.Image // list of tiles (0..Width*Height)
}

// Item return the tile at (linear) position i
func (t *Tiles) Item(i int) *ebiten.Image {
	if i < 0 || i >= len(t.List) {
		return nil
	}

	return t.List[i]
}

// At return the tile at the coordinates x, y
func (t *Tiles) At(x, y int) *ebiten.Image {
	if x < 0 || x >= t.Columns || y < 0 || y >= t.Rows {
		return nil
	}

	return t.List[x+y*t.Width]
}

// ReadTiles read the tiles from a png file
func ReadTiles(r io.Reader, nx, ny int) (*Tiles, error) {
	return ReadTilesScaledTo(r, nx, ny, 0, 0)
}

func ReadTilesScaledTo(r io.Reader, nx, ny int, tw, th int) (*Tiles, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	if iw%nx != 0 || ih%ny != 0 {
		return nil, errors.New("invalid cells image dimension")
	}

	ebimg := ebiten.NewImageFromImage(img)

	if tw > 0 && th > 0 {
		sw, sh := float64(tw)/float64(iw/nx), float64(th)/float64(ih/ny)

		scaled := ebiten.NewImage(iw, ih)

		var op ebiten.DrawImageOptions
		op.GeoM.Scale(sw, sh)

		scaled.DrawImage(ebimg, &op)
		ebimg = scaled
	} else {
		tw, th = iw/nx, ih/ny
	}

	p := image.Rect(0, 0, tw, th)

	var tiles = Tiles{Width: tw, Height: th, Rows: ny, Columns: nx}

	y := 0

	for v := 0; v < ny; v++ {
		for h := 0; h < nx; h++ {
			tile := ebimg.SubImage(p).(*ebiten.Image)
			tiles.List = append(tiles.List, tile)
			p = p.Add(image.Pt(tw, 0))
		}

		y += th
		p = image.Rect(0, y, tw, y+th)
	}

	return &tiles, nil
}
