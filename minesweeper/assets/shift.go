package main

import (
	"flag"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func getImage(fname string) (image.Image, string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, fname, err
	}

	defer f.Close()

	ima, _, err := image.Decode(f)
	if err != nil {
		return nil, fname, err
	}

	fname = filepath.Base(fname)
	log.Println(fname, "width:", ima.Bounds().Dx(), "height:", ima.Bounds().Dy())

	return ima, fname, nil

}

func main() {
	n := flag.Int("n", 0, "number of sprites")
	sp := flag.Int("space", 1, "space between sprites")

	flag.Parse()

	if flag.NArg() == 0 || *n <= 0 || *sp <= 0 {
		log.Fatal("usage: shift -n=nsprites [-sp=space] <image>...")
	}

	for _, fn := range flag.Args() {
		ima, _, err := getImage(fn)
		if err != nil {
			log.Println(fn, err)
			continue
		}

		w := ima.Bounds().Dx()
		h := ima.Bounds().Dy()

		sw := (w - (*sp * (*n - 1))) / *n
		log.Println("sprite width:", sw, "height:", h)
		log.Println("image width:", sw**n, "height:", h)

		nima := image.NewNRGBA(image.Rect(0, 0, sw**n, h))

		sp := image.Rect(0, 0, sw, h)
		p := image.Pt(0, 0)

		for i := 0; i < *n; i++ {
			draw.Draw(nima, sp, ima, p, draw.Over)
			sp = sp.Add(image.Pt(sw, 0))
			p = p.Add(image.Pt(sw+1, 0))
		}

		f, err := os.Create(fn)
		if err != nil {
			log.Println(err)
		}

		png.Encode(f, nima)
		f.Close()
	}
}
