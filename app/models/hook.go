package models

import (
	"math/rand"
	"strings"
	"time"

	"github.com/martinlindhe/base36"

	"github.com/dotzero/hooks/app/favicon"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Hook is a hook model
type Hook struct {
	Name     string     `json:"name"`
	Secret   string     `json:"secret"`
	Private  bool       `json:"private"`
	Color    [4]uint8   `json:"color"`
	Created  time.Time  `json:"time"`
	Requests []*Request `json:"-"`
}

// NewHook returns a new hook model
func NewHook(private bool) *Hook {
	rgba := favicon.RandomRGBA()

	hook := &Hook{
		Name:    tinyID(),
		Created: time.Now().UTC(),
		Color:   [4]uint8{rgba.R, rgba.G, rgba.B, rgba.A},
	}

	if private {
		hook.Secret = tinyID()
		hook.Private = true
	}

	return hook
}

func tinyID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b) // nolint:gosec

	return strings.ToLower(base36.EncodeBytes(b))
}
