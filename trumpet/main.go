package main

import (
	"bytes"
	_ "embed"
        "flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/raff/ebi-games/util"
)

const (
	sampleRate = 8000
)

type Note int

const (
	O3_E Note = iota
	O3_F
	O3_Fx
	O3_G
	O3_Gx
	O3_A
	O3_Ax
	O3_B

	O4_C
	O4_Cx
	O4_D
	O4_Dx
	O4_E
	O4_F
	O4_Fx
	O4_G
	O4_Gx
	O4_A
	O4_Ax
	O4_B

	O5_C
	O5_Cx
	O5_D
	O5_Dx
	O5_E
	O5_F
	O5_Fx
	O5_G
	O5_Gx
	O5_A
	O5_Ax
	O5_B

	O6_C
	O6_Cx
	O6_D
	O6_Dx

	// Partials
	P1 = 0x100
	P2 = 0x200
	P3 = 0x400
	P4 = 0x800
	P5 = 0x1000

	// Valves
	V0   = 0b000 // open
	V1   = 0b100 // first valve
	V2   = 0b010 // second
	V3   = 0b001 // third
	V12  = V1 | V2
	V13  = V1 | V3
	V23  = V2 | V3
	V123 = V1 | V2 | V3
)

var (
	//go:embed assets/trumpet_valves.png
	trumpetPng []byte

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
	notes        map[Note]*audio.Player
	knotes       map[ebiten.Key]Note
	tnotes       map[int]Note
	tvalves      = 0

	tiles *util.Tiles

	screenWidth  int
	screenHeight int

        playAudio = true
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
	notes = map[Note]*audio.Player{
		O3_E:  newWavPlayer(n00), // E
		O3_F:  newWavPlayer(n01), // F
		O3_Fx: newWavPlayer(n02), // F#
		O3_G:  newWavPlayer(n03), // G
		O3_Gx: newWavPlayer(n04), // G#
		O3_A:  newWavPlayer(n05), // A
		O3_Ax: newWavPlayer(n06), // A# / Bb
		O3_B:  newWavPlayer(n07), // B
		O4_C:  newWavPlayer(n08), // C4
		O4_Cx: newWavPlayer(n09), // C#
		O4_D:  newWavPlayer(n10), // D
		O4_Dx: newWavPlayer(n11), // D# / Eb

		O4_E:  newWavPlayer(n12), // E
		O4_F:  newWavPlayer(n13), // F
		O4_Fx: newWavPlayer(n14), // F#
		O4_G:  newWavPlayer(n15), // G
		O4_Gx: newWavPlayer(n16), // G#
		O4_A:  newWavPlayer(n17), // A
		O4_Ax: newWavPlayer(n18), // A# / Bb
		O4_B:  newWavPlayer(n19), // B
		O5_C:  newWavPlayer(n20), // C5
		O5_Cx: newWavPlayer(n21), // C#
		O5_D:  newWavPlayer(n22), // D
		O5_Dx: newWavPlayer(n23), // D# / Eb

		O5_E:  newWavPlayer(n24), // E
		O5_F:  newWavPlayer(n25), // F
		O5_Fx: newWavPlayer(n26), // F#
		O5_G:  newWavPlayer(n27), // G
		O5_Gx: newWavPlayer(n28), // G#
		O5_A:  newWavPlayer(n29), // A
		O5_Ax: newWavPlayer(n30), // A# / Bb
		O5_B:  newWavPlayer(n31), // B
		O6_C:  newWavPlayer(n32), // C6
		O6_Cx: newWavPlayer(n33), // C#
		O6_D:  newWavPlayer(n34), // D
		O6_Dx: newWavPlayer(n35), // D#
	}

	knotes = map[ebiten.Key]Note{
		ebiten.Key1:     O3_E,
		ebiten.Key2:     O3_F,
		ebiten.Key3:     O3_Fx,
		ebiten.Key4:     O3_G,
		ebiten.Key5:     O3_Gx,
		ebiten.Key6:     O3_A,
		ebiten.Key7:     O3_Ax,
		ebiten.Key8:     O3_B,
		ebiten.Key9:     O4_C,
		ebiten.Key0:     O4_Cx,
		ebiten.KeyMinus: O4_D,
		ebiten.KeyEqual: O4_Dx,

		ebiten.KeyQ:            O4_E,
		ebiten.KeyW:            O4_F,
		ebiten.KeyE:            O4_Fx,
		ebiten.KeyR:            O4_G,
		ebiten.KeyT:            O4_Gx,
		ebiten.KeyY:            O4_A,
		ebiten.KeyU:            O4_Ax,
		ebiten.KeyI:            O4_B,
		ebiten.KeyO:            O5_C,
		ebiten.KeyP:            O5_Cx,
		ebiten.KeyBracketLeft:  O5_D,
		ebiten.KeyBracketRight: O5_Dx,

		ebiten.KeyA:         O5_E,
		ebiten.KeyS:         O5_F,
		ebiten.KeyD:         O5_Fx,
		ebiten.KeyF:         O5_G,
		ebiten.KeyG:         O5_Gx,
		ebiten.KeyH:         O5_A,
		ebiten.KeyJ:         O5_Ax,
		ebiten.KeyK:         O5_B,
		ebiten.KeyL:         O6_C,
		ebiten.KeySemicolon: O6_Cx,
		ebiten.KeyQuote:     O6_D,
		ebiten.KeyEnter:     O6_Dx,
	}

	tnotes = map[int]Note{
		P1 | V123: O3_E,
		P1 | V13:  O3_F,
		P1 | V23:  O3_Fx,
		P1 | V12:  O3_G,
		P1 | V1:   O3_Gx,
		P1 | V2:   O3_A,
		P1 | V0:   O3_Ax,

		P2 | V123: O3_B,
		P2 | V13:  O4_C,
		P2 | V23:  O4_Cx,
		P2 | V12:  O4_D,
		P2 | V1:   O4_Dx,
		P2 | V2:   O4_E,
		P2 | V0:   O4_F,

		P3 | V23: O4_Fx,
		P3 | V12: O4_G,
		P3 | V1:  O4_Gx,
		P3 | V2:  O4_A,
		P3 | V0:  O4_Ax,

		P4 | V12: O4_B,
		P4 | V1:  O5_C,
		P4 | V2:  O5_Cx,
		P4 | V0:  O5_D,

		P5 | V1: O5_Dx,
		P5 | V2: O5_E,
		P5 | V0: O5_F,
	}

	if t, err := util.ReadTiles(bytes.NewBuffer(trumpetPng), 2, 4); err == nil {
		tiles = t
	} else {
		log.Fatalf("%+v", err)
	}

	screenWidth = tiles.Width
	screenHeight = tiles.Height
}

func pplayNote(n Note, play bool) {
        if !playAudio {
                return
        }

	p := notes[n]
	if play {
		p.Rewind()
		p.Play()
	} else {
		p.Pause()
	}
}

type Game struct {
	redraw bool
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	ima := tiles.Item(tvalves & V123)
	screen.DrawImage(ima, &ebiten.DrawImageOptions{})

	g.redraw = false
}

func (g *Game) Update() error {
	for k, v := range knotes {
		if inpututil.IsKeyJustReleased(k) {
			pplayNote(v, false)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	for k, v := range knotes {
		if inpututil.IsKeyJustPressed(k) {
			pplayNote(v, true)
		}
	}

	valves := tvalves & V123

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		valves |= V1
	} else {
		valves &= V23
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		valves |= V2
	} else {
		valves &= V13
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		valves |= V3
	} else {
		valves &= V12
	}

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyShift):
		valves |= P5
	case ebiten.IsKeyPressed(ebiten.KeyControl):
		valves |= P4
	case ebiten.IsKeyPressed(ebiten.KeyAlt):
		valves |= P3
	case ebiten.IsKeyPressed(ebiten.KeyMeta):
		valves |= P2
	case ebiten.IsKeyPressed(ebiten.KeySpace):
		valves |= P1
	}

	if valves != tvalves {
		if n, ok := tnotes[tvalves]; ok {
			pplayNote(n, false)
		}

		tvalves = valves
		log.Println("valves", valves)

		if n, ok := tnotes[tvalves]; ok {
			pplayNote(n, true)
		}

		g.redraw = true
	}

	return nil
}

func main() {
        flag.BoolVar(&playAudio, "audio", playAudio, "play notes")
        flag.Parse()

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Trumpetine")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(&Game{redraw: true}); err != nil {
		log.Fatal(err)
	}
}
