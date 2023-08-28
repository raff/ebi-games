// ## go:build !ios && !android && !js

package main

import (
	"bytes"
	"log"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	AudioHit  = 0
	AudioMiss = 1
)

var (
	//go:embed assets/ping.wav
	wavPing []byte

	//go:embed assets/buzz.wav
	wavBuzz []byte

	audioPlayers []*audio.Player
)

func audioInit() {
	audioContext := audio.NewContext(8000)

	wbits, err := wav.Decode(audioContext, bytes.NewReader(wavPing))
	if err != nil {
		log.Fatalf("%+v", err)
	}

	p, err := audioContext.NewPlayer(wbits)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	audioPlayers = append(audioPlayers, p)

	wbits, err = wav.Decode(audioContext, bytes.NewReader(wavBuzz))
	if err != nil {
		log.Fatalf("%+v", err)
	}

	p, err = audioContext.NewPlayer(wbits)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	audioPlayers = append(audioPlayers, p)
}

func audioPlay(aid int) {
	lp := len(audioPlayers) - 1

	if aid < 0 || aid > lp {
		return
	}

	p := audioPlayers[aid]
	go func() {
		p.Rewind()
		p.Play()
	}()
}
