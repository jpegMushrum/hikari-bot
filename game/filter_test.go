package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterIsJapanese(t *testing.T) {
	assert.True(t, isJapanese("ー"))
	assert.True(t, isJapanese("ヘヘ"))
	assert.True(t, isJapanese("着る"))
	assert.True(t, isJapanese("るす"))
	assert.False(t, isJapanese("asdasd"))
	assert.False(t, isJapanese("привет"))
}

func TestFiltersIsJapSuitable(t *testing.T) {
	assert.False(t, isJapSuitable("    "))
	assert.False(t, isJapSuitable(""))
	assert.True(t, isJapSuitable("着る"))
}

func TestFilterIsSmall(t *testing.T) {
	assert.True(t, isSmall('ォ'))
	assert.True(t, isSmall('ぁ'))
	assert.False(t, isSmall('ア'))
	assert.False(t, isSmall('え'))
}

func TestFilterToBigKana(t *testing.T) {
	assert.Equal(t, 'オ', toBigKana('ォ'))
	assert.Equal(t, 'や', toBigKana('ゃ'))
	assert.Equal(t, 'い', toBigKana('ぃ'))
}

func TestFilterGetFirstKana(t *testing.T) {
	assert.Equal(t, 'へ', getFirstKana("へんたい"))
	assert.Equal(t, 'き', getFirstKana("キス"))
	assert.Equal(t, 'ら', getFirstKana("ラ"))
	assert.Equal(t, int32(0), getFirstKana("ー"))
	assert.Equal(t, 'あ', getFirstKana("ーア"))
}

func TestFilterGetLastKana(t *testing.T) {
	assert.Equal(t, 'い', getLastKana("へんたい"))
	assert.Equal(t, 'す', getLastKana("キス"))
	assert.Equal(t, 'ら', getLastKana("ラ"))
	assert.Equal(t, 'ー', getLastKana("スキー"))
	assert.NotEqual(t, 'ょ', getLastKana("しゅしょ"))
	assert.Equal(t, 'よ', getLastKana("しゅしょ"))
}

func TestFilterEqualities(t *testing.T) {
	assert.Equal(t, getFirstKana("ラ"), getLastKana("ラ"))
	assert.Equal(t, getFirstKana("おにぎり"), getLastKana("ラジオ"))
	assert.Equal(t, getFirstKana("おにぎり"), getLastKana("ラジォ"))
	assert.Equal(t, getFirstKana("オにぎり"), getLastKana("ラジぉ"))
	assert.Equal(t, getFirstKana("じごく"), getLastKana("ジジ"))
	assert.Equal(t, getFirstKana("ぱら"), getLastKana("パパ"))
}
