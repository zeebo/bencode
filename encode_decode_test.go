package bencode

import (
	"errors"
	"time"
)

type myBoolType bool

// MarshalBencode implements Marshaler.MarshalBencode
func (mbt myBoolType) MarshalBencode() ([]byte, error) {
	var c string
	if mbt {
		c = "y"
	} else {
		c = "n"
	}

	return EncodeBytes(c)
}

// UnmarshalBencode implements Unmarshaler.UnmarshalBencode
func (mbt *myBoolType) UnmarshalBencode(b []byte) error {
	var str string
	err := DecodeBytes(b, &str)
	if err != nil {
		return err
	}

	switch str {
	case "y":
		*mbt = true
	case "n":
		*mbt = false
	default:
		err = errors.New("invalid myBoolType")
	}

	return err
}

type myTimeType struct {
	time.Time
}

// MarshalBencode implements Marshaler.MarshalBencode
func (mtt myTimeType) MarshalBencode() ([]byte, error) {
	return EncodeBytes(mtt.Time.Unix())
}

// UnmarshalBencode implements Unmarshaler.UnmarshalBencode
func (mtt *myTimeType) UnmarshalBencode(b []byte) error {
	var epoch int64
	err := DecodeBytes(b, &epoch)
	if err != nil {
		return err
	}

	mtt.Time = time.Unix(epoch, 0)
	return nil
}

type errorMarshalType struct{}

// MarshalBencode implements Marshaler.MarshalBencode
func (emt errorMarshalType) MarshalBencode() ([]byte, error) {
	return nil, errors.New("oops")
}

// UnmarshalBencode implements Unmarshaler.UnmarshalBencode
func (emt errorMarshalType) UnmarshalBencode([]byte) error {
	return errors.New("oops")
}
