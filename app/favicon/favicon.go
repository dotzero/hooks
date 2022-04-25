package favicon

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"sync"
	"time"
)

var once sync.Once

// Favicon provides a rectangular random colored image that can be used as a favicon
type Favicon struct {
	*image.RGBA
}

const (
	maxComponent = 255
)

// New is a favicon constuctor
func New(options ...Option) *Favicon {
	opts := Options{
		Height: 16,
		Width:  16,
		Color:  RandomRGBA(),
	}

	for _, o := range options {
		o(&opts)
	}

	favicon := &Favicon{image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))}

	draw.Draw(
		favicon,
		favicon.Bounds(),
		&image.Uniform{opts.Color},
		image.Point{},
		draw.Src,
	)

	return favicon
}

// String implements Stringer interface
func (f *Favicon) String() string {
	var buf bytes.Buffer

	if err := png.Encode(&buf, f); err != nil {
		panic(err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/png;base64," + encoded
}

// RandomRGBA returns random coloder RGBA
func RandomRGBA() color.RGBA {
	return color.RGBA{randomColor(10, 5), randomColor(10, 5), randomColor(10, 5), maxComponent}
}

func randomColor(factor int, min int) uint8 {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})

	max := maxComponent / factor

	return uint8((rand.Intn(max-min) + min) * factor) // nolint:gosec
}
