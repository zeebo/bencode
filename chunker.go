package bencode

import (
	"io"
	"bufio"
	"os"
)

type chunker struct {
	r    io.Reader
	buf  []byte
	errd bool
}

func newChunker(r io.Reader) *chunker {
	return &chunker{r, false}
}

func (c *chunker) nextValue() (string, os.Error) {
	//read from r
}
