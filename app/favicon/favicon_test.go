package favicon

import (
	"image/color"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	f := New()

	assert.Equal(t, 16, f.Bounds().Size().X)
	assert.Equal(t, 16, f.Bounds().Size().Y)
}

func TestOpts(t *testing.T) {
	f := New(WithSize(10, 20), WithColor(color.Black))

	assert.Equal(t, 10, f.Bounds().Size().X)
	assert.Equal(t, 20, f.Bounds().Size().Y)

	r1, g1, b1, a1 := color.Black.RGBA()
	r2, g2, b2, a2 := f.At(1, 1).RGBA()

	assert.Equal(t, r1, r2)
	assert.Equal(t, g1, g2)
	assert.Equal(t, b1, b2)
	assert.Equal(t, a1, a2)
}

func TestString(t *testing.T) {
	str := New().String()

	assert.True(t, strings.HasPrefix(str, "data:image/png;base64,"))
	assert.True(t, len(str) > len("data:image/png;base64,"))
}
