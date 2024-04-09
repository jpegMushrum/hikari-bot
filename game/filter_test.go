package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKana(t *testing.T) {
	assert.Equal(t, GetFirstKana("へんたい"), 'へ')
	assert.Equal(t, GetFirstKana("キス"), 'キ')
}
