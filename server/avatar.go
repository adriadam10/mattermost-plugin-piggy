package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
)

// The CC0 sprite sheet by Natural_Privateer (see README); the webapp draws from the same file.
//
//go:embed assets/piggy_look_anim-sheet.png
var spriteSheet []byte

const (
	frameSize   = 32
	avatarScale = 4
)

// The sheet is peach; the piggy is pink. Both sides map the sheet's three colours the same way.
var pinkPalette = map[color.RGBA]color.RGBA{
	{238, 195, 154, 255}: {255, 183, 197, 255},
	{217, 160, 102, 255}: {232, 128, 160, 255},
	{143, 86, 59, 255}:   {138, 59, 82, 255},
}

var avatarBackground = color.RGBA{255, 240, 214, 255}

// avatarPNG is the first frame of the sheet in pink, scaled up without smoothing.
func avatarPNG() ([]byte, error) {
	sheet, err := png.Decode(bytes.NewReader(spriteSheet))
	if err != nil {
		return nil, err
	}
	size := frameSize * avatarScale
	avatar := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			// The pig stands on rows 6-27 of its frame; nudging it down 2 rows centres it.
			src := color.RGBAModel.Convert(sheet.At(x/avatarScale, y/avatarScale-2)).(color.RGBA)
			if pink, ok := pinkPalette[src]; ok {
				avatar.Set(x, y, pink)
			} else {
				avatar.Set(x, y, avatarBackground)
			}
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, avatar); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
