package bencode

import (
	"io"
	"bufio"
	"os"
	"strconv"
)

type chunker struct {
	r    *bufio.Reader
	errd bool
}

func NewChunker(r io.Reader) *chunker {
	return &chunker{bufio.NewReader(r), false}
}

func (c *chunker) Next() (string, os.Error) { return c.nextValue() }

func (c *chunker) nextValue() (string, os.Error) {
	//peek a byte and figure out
	b, err := c.r.Peek(1)
	if err != nil {
		return "", err
	}
	switch b[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return c.nextString()
	case 'i':
		return c.nextInt()
	case 'l':
		return c.nextList()
	case 'd':
		return c.nextDict()
	}

	return "", os.NewError("Unexpected delimiter")
}

func (c *chunker) nextString() (string, os.Error) {
	//read until the :
	num, err := c.r.ReadString(':')
	if err != nil {
		return "", err
	}

	n, err := strconv.Atoi64(num[:len(num)-1])
	if err != nil {
		return "", err
	}

	buf := make([]byte, n)
	var p int64
	for p != n {
		nr, err := c.r.Read(buf[p:])
		if err != nil {
			return "", err
		}
		p += int64(nr)
	}

	return num + string(buf), nil
}

func (c *chunker) nextInt() (string, os.Error) {
	_, err := c.r.ReadByte()
	if err != nil {
		return "", err
	}

	val, err := c.r.ReadString('e')
	if err != nil {
		return "", err
	}
	return "i" + val, nil
}

func (c *chunker) nextList() (string, os.Error) {
	//read off the beginning delimiter
	_, err := c.r.ReadByte()
	if err != nil {
		return "", err
	}

	buf := make([]byte, 0)
	for {
		bs, err := c.r.Peek(1)
		if err != nil {
			return "", err
		}
		//peek an e
		if bs[0] == 'e' {
			break
		}

		nv, err := c.nextValue()
		if err != nil {
			return "", err
		}

		buf = append(buf, []byte(nv)...)
	}

	return "l" + string(buf), nil
}

func (c *chunker) nextDict() (string, os.Error) {
	_, err := c.r.ReadByte()
	if err != nil {
		return "", err
	}

	buf := make([]byte, 0)
	for {
		bs, err := c.r.Peek(1)
		if err != nil {
			return "", err
		}

		if bs[0] == 'e' {
			break
		}

		switch bs[0] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return "", os.NewError("Key is not a string")
		}

		ns, err := c.nextString()
		if err != nil {
			return "", err
		}

		buf = append(buf, []byte(ns)...)

		nv, err := c.nextValue()
		if err != nil {
			return "", err
		}

		buf = append(buf, []byte(nv)...)
	}

	return "d" + string(buf), nil
}