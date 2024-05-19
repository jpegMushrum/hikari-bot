package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilters(t *testing.T) {
	assert.True(t, IsJapanese("ー"))
	assert.True(t, IsJapanese("ヘヘ"))
	assert.True(t, IsJapanese("着る"))
	assert.True(t, IsJapanese("るす"))
	assert.False(t, IsJapanese("asdasd"))
	assert.False(t, IsJapanese("привет"))
	assert.False(t, IsJapSuitable("    "))
	assert.False(t, IsJapSuitable(""))
	assert.True(t, IsJapSuitable("着る"))
	assert.True(t, IsSmall('ォ'))
	assert.True(t, IsSmall('ぁ'))
	assert.False(t, IsSmall('ア'))
	assert.False(t, IsSmall('え'))
	assert.Equal(t, ToBigKana('ォ'), 'オ')
	assert.Equal(t, ToBigKana('ゃ'), 'や')
	assert.Equal(t, ToBigKana('ぃ'), 'い')
	assert.Equal(t, GetFirstKana("へんたい"), 'へ')
	assert.Equal(t, GetFirstKana("キス"), 'き')
	assert.Equal(t, GetFirstKana("ラ"), 'ら')
	assert.Equal(t, GetFirstKana("ー"), 'ー')
	assert.Equal(t, GetLastKana("へんたい"), 'い')
	assert.Equal(t, GetLastKana("キス"), 'す')
	assert.Equal(t, GetLastKana("ラ"), 'ら')
	assert.Equal(t, GetLastKana("スキー"), 'ー')
	assert.NotEqual(t, GetLastKana("しゅしょ"), 'ょ')
	assert.Equal(t, GetLastKana("しゅしょ"), 'よ')
	assert.Equal(t, GetFirstKana("ラ"), GetLastKana("ラ"))
	assert.Equal(t, GetLastKana("ラジオ"), GetFirstKana("おにぎり"))
	assert.Equal(t, GetLastKana("ラジォ"), GetFirstKana("おにぎり"))
	assert.Equal(t, GetLastKana("ラジぉ"), GetFirstKana("オにぎり"))
	assert.Equal(t, GetLastKana("ジジ"), GetFirstKana("じごく"))
	assert.Equal(t, GetLastKana("パパ"), GetFirstKana("ぱら"))
}
