package asciitrie

import "github.com/odeke-em/dts/trie"

func New() *trie.Trie {
	return trie.New(trie.AsciiAlphabet)
}
