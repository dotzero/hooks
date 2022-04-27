package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHook(t *testing.T) {
	model := NewHook(true)

	assert.NotEmpty(t, model.Name)
	assert.True(t, model.Private)
	assert.NotEmpty(t, model.Secret)
}
