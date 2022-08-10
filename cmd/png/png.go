package main

import (
	"image/png"
	"log"
	"os"
	"png/mbezuidenhout/kindleland"

	"github.com/fogleman/gg"
)

func main() {
	fb, err := kindleland.NewFrameBuffer("/dev/fb0", 600, 800)
	if err != nil {
		panic(err)
	}

	img := fb.Image()

	f, err := os.Create("screengrab.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	ctx, err := gg.LoadPNG("image.png")
	if err != nil {
		panic(err)
	}

	fb.ClearScreen()

	err = fb.ApplyImage(ctx)
	if err != nil {
		panic(err)
	}

	err = fb.UpdateScreen()
	if err != nil {
		panic(err)
	}
}
