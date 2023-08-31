package main

import (
	"bytes"
	_ "embed"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	sampleRate   = 8000
	screenWidth  = 100
	screenHeight = 100
)

var (
	//go:embed assets/_1-E_3.wav
	n00 []byte

	//go:embed assets/_2-F_3.wav
	n01 []byte

	//go:embed assets/_3-F+3.wav
	n02 []byte

	//go:embed assets/_4-G_3.wav
	n03 []byte

	//go:embed assets/_5-G+3.wav
	n04 []byte

	//go:embed assets/_6-A_3.wav
	n05 []byte

	//go:embed assets/_7-A+3.wav
	n06 []byte

	//go:embed assets/_8-B_3.wav
	n07 []byte

	//go:embed assets/01-C_4.wav
	n08 []byte

	//go:embed assets/02-C+4.wav
	n09 []byte

	//go:embed assets/03-D_4.wav
	n10 []byte

	//go:embed assets/04-D+4.wav
	n11 []byte

	//go:embed assets/05-E_4.wav
	n12 []byte

	//go:embed assets/06-F_4.wav
	n13 []byte

	//go:embed assets/07-F+4.wav
	n14 []byte

	//go:embed assets/08-G_4.wav
	n15 []byte

	//go:embed assets/09-G+4.wav
	n16 []byte

	//go:embed assets/10-A_4.wav
	n17 []byte

	//go:embed assets/11-A+4.wav
	n18 []byte

	//go:embed assets/12-B_4.wav
	n19 []byte

	//go:embed assets/13-C_5.wav
	n20 []byte

	//go:embed assets/14-C+5.wav
	n21 []byte

	//go:embed assets/15-D_5.wav
	n22 []byte

	//go:embed assets/16-D+5.wav
	n23 []byte

	//go:embed assets/17-E_5.wav
	n24 []byte

	//go:embed assets/18-F_5.wav
	n25 []byte

	//go:embed assets/19-F+5.wav
	n26 []byte

	//go:embed assets/20-G_5.wav
	n27 []byte

	//go:embed assets/21-G+5.wav
	n28 []byte

	//go:embed assets/22-A_5.wav
	n29 []byte

	//go:embed assets/23-A+5.wav
	n30 []byte

	//go:embed assets/24-B_5.wav
	n31 []byte

	//go:embed assets/25-C_6.wav
	n32 []byte

	//go:embed assets/26-C+6.wav
	n33 []byte

	//go:embed assets/27-D_6.wav
	n34 []byte

	//go:embed assets/28-D+6.wav
	n35 []byte

	audioContext = audio.NewContext(sampleRate)
	notes        map[ebiten.Key]*audio.Player
	tnotes       map[int]*audio.Player
	tvalves      = 0
)

func newWavPlayer(b []byte) *audio.Player {
	wbits, err := wav.Decode(audioContext, bytes.NewReader(b))
	if err != nil {
		log.Fatalf("%+v", err)
	}

	p, err := audioContext.NewPlayer(wbits)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	return p
}

func init() {
	notes = map[ebiten.Key]*audio.Player{
		ebiten.Key1:     newWavPlayer(n00),
		ebiten.Key2:     newWavPlayer(n01),
		ebiten.Key3:     newWavPlayer(n02),
		ebiten.Key4:     newWavPlayer(n03),
		ebiten.Key5:     newWavPlayer(n04),
		ebiten.Key6:     newWavPlayer(n05),
		ebiten.Key7:     newWavPlayer(n06), // Bb
		ebiten.Key8:     newWavPlayer(n07),
		ebiten.Key9:     newWavPlayer(n08), // C4
		ebiten.Key0:     newWavPlayer(n09),
		ebiten.KeyMinus: newWavPlayer(n10),
		ebiten.KeyEqual: newWavPlayer(n11),

		ebiten.KeyQ:            newWavPlayer(n12),
		ebiten.KeyW:            newWavPlayer(n13),
		ebiten.KeyE:            newWavPlayer(n14),
		ebiten.KeyR:            newWavPlayer(n15),
		ebiten.KeyT:            newWavPlayer(n16),
		ebiten.KeyY:            newWavPlayer(n17),
		ebiten.KeyU:            newWavPlayer(n18),
		ebiten.KeyI:            newWavPlayer(n19),
		ebiten.KeyO:            newWavPlayer(n20), // C5
		ebiten.KeyP:            newWavPlayer(n21),
		ebiten.KeyBracketLeft:  newWavPlayer(n22),
		ebiten.KeyBracketRight: newWavPlayer(n23),

		ebiten.KeyA:         newWavPlayer(n24),
		ebiten.KeyS:         newWavPlayer(n25),
		ebiten.KeyD:         newWavPlayer(n26),
		ebiten.KeyF:         newWavPlayer(n27),
		ebiten.KeyG:         newWavPlayer(n28),
		ebiten.KeyH:         newWavPlayer(n29),
		ebiten.KeyJ:         newWavPlayer(n30),
		ebiten.KeyK:         newWavPlayer(n31),
		ebiten.KeyL:         newWavPlayer(n32), // C6
		ebiten.KeySemicolon: newWavPlayer(n33),
		ebiten.KeyQuote:     newWavPlayer(n34),
		ebiten.KeyEnter:     newWavPlayer(n35),
	}

	tnotes = map[int]*audio.Player{
		0x100: notes[ebiten.Key7],
		0x107: notes[ebiten.Key8],
		0x105: notes[ebiten.Key9],
		0x103: notes[ebiten.Key0],
		0x106: notes[ebiten.KeyMinus],
		0x101: notes[ebiten.KeyMinus],
		0x104: notes[ebiten.KeyEqual],
		0x102: notes[ebiten.KeyQ],
	}
}

type Game struct {
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0x80, 0xc0, 0xff})
	//ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) Update() error {
	for k, v := range notes {
		if inpututil.IsKeyJustReleased(k) {
			v.Pause()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	for k, v := range notes {
		if inpututil.IsKeyJustPressed(k) {
			v.Rewind()
			v.Play()
		}
	}

	valves := tvalves

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		valves = (valves & 7) | 1
	} else {
		valves = valves & 6
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		valves = (valves & 7) | 2
	} else {
		valves = valves & 5
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		valves = (valves & 7) | 4
	} else {
		valves = valves & 3
	}

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyMeta):
		valves |= 0x10
	case ebiten.IsKeyPressed(ebiten.KeyAlt):
		valves |= 0x20
	case ebiten.IsKeyPressed(ebiten.KeyControl):
		valves |= 0x40
	case ebiten.IsKeyPressed(ebiten.KeyShift):
		valves |= 0x80
	case ebiten.IsKeyPressed(ebiten.KeySpace):
		valves |= 0x100
	}

	if valves != tvalves {
		if p, ok := tnotes[tvalves]; ok {
			log.Println("pause", tvalves, p)
			p.Pause()
		}

		tvalves = valves
		log.Println("valves", valves)

		if p, ok := tnotes[tvalves]; ok {
			log.Println("play", tvalves, p)

			p.Rewind()
			p.Play()
		}
	}

	return nil
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Trumpetine")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
