package favicon

import (
	"image/color"
)

// Options is favicon options
type Options struct {
	Width  int
	Height int
	Color  color.Color
}

// Option function type
type Option func(o *Options)

// WithSize setting up a favicon with a specific size
func WithSize(width int, height int) Option {
	return func(o *Options) {
		o.Width = width
		o.Height = height
	}
}

// WithColor setting up a favicon with a specific color
func WithColor(c color.Color) Option {
	return func(o *Options) {
		o.Color = c
	}
}
