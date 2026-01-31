package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilters(t *testing.T) {
	assert.True(t, isJapanese("ー"))
	assert.True(t, isJapanese("ヘヘ"))
	assert.True(t, isJapanese("着る"))
	assert.True(t, isJapanese("るす"))
	assert.False(t, isJapanese("asdasd"))
	assert.False(t, isJapanese("привет"))
	assert.False(t, isJapSuitable("    "))
	assert.False(t, isJapSuitable(""))
	assert.True(t, isJapSuitable("着る"))
	assert.True(t, isSmall('ォ'))
	assert.True(t, isSmall('ぁ'))
	assert.False(t, isSmall('ア'))
	assert.False(t, isSmall('え'))
	assert.Equal(t, toBigKana('ォ'), 'オ')
	assert.Equal(t, toBigKana('ゃ'), 'や')
	assert.Equal(t, toBigKana('ぃ'), 'い')
	assert.Equal(t, getFirstKana("へんたい"), 'へ')
	assert.Equal(t, getFirstKana("キス"), 'き')
	assert.Equal(t, getFirstKana("ラ"), 'ら')
	assert.Equal(t, getFirstKana("ー"), 'ー')
	assert.Equal(t, getLastKana("へんたい"), 'い')
	assert.Equal(t, getLastKana("キス"), 'す')
	assert.Equal(t, getLastKana("ラ"), 'ら')
	assert.Equal(t, getLastKana("スキー"), 'ー')
	assert.NotEqual(t, getLastKana("しゅしょ"), 'ょ')
	assert.Equal(t, getLastKana("しゅしょ"), 'よ')
	assert.Equal(t, getFirstKana("ラ"), getLastKana("ラ"))
	assert.Equal(t, getLastKana("ラジオ"), getFirstKana("おにぎり"))
	assert.Equal(t, getLastKana("ラジォ"), getFirstKana("おにぎり"))
	assert.Equal(t, getLastKana("ラジぉ"), getFirstKana("オにぎり"))
	assert.Equal(t, getLastKana("ジジ"), getFirstKana("じごく"))
	assert.Equal(t, getLastKana("パパ"), getFirstKana("ぱら"))
}
