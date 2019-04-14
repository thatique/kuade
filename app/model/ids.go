package model

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/jsonpb"
)

// ID is identifier for unique identifier
type ID uint64

// NewIDFromString create ID from string
func NewIDFromString(s string) (ID, error) {
	if len(s) > 16 {
		return ID(0), fmt.Errorf("SpanID cannot be longer than 16 hex characters: %s", s)
	}
	id, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return ID(0), err
	}
	return ID(id), nil
}

func (id ID) String() string {
	return fmt.Sprintf("%x", uint64(s))
}

// Base63 return Base32 representation of this ID
func (id ID) Base63() string {
	var b [11]byte
	encode(&b, uint64(id))
	return string(b[:])
}

// Size returns the size of this datum in protobuf. It is always 8 bytes.
func (id *ID) Size() int {
	return 8
}

// MarshalTo converts ID into a binary representation. Called by protobuf serialization.
func (id *ID) MarshalTo(data []byte) (n int, err error) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(*id))
	return marshalBytes(data, b[:])
}

// Unmarshal inflates ID from a binary representation. Called by protobuf serialization.
func (id *ID) Unmarshal(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("buffer is too short")
	}
	*id = ID(binary.BigEndian.Uint64(data))
	return nil
}

// MarshalJSON converts span id into a base64 string enclosed in quotes.
// Used by protobuf JSON serialization.
// Example: {1} => "AAAAAAAAAAE=".
func (id ID) MarshalJSON() ([]byte, error) {
	var b [8]byte
	id.MarshalTo(b[:]) // can only error on incorrect buffer size
	v := make([]byte, 12+2)
	base64.StdEncoding.Encode(v[1:13], b[:])
	v[0], v[13] = '"', '"'
	return v, nil
}

// UnmarshalJSON inflates id from base64 string, possibly enclosed in quotes.
// User by protobuf JSON serialization.
//
// There appears to be a bug in gogoproto, as this function is only called for numeric values.
// https://github.com/gogo/protobuf/issues/411#issuecomment-393856837
func (id *ID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if l := len(str); l > 2 && str[0] == '"' && str[l-1] == '"' {
		str = str[1 : l-1]
	}
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return fmt.Errorf("cannot unmarshal ID from string '%s': %v", string(data), err)
	}
	return id.Unmarshal(b)
}

// UnmarshalJSONPB inflates id from base64 string, possibly enclosed in quotes.
// User by protobuf JSON serialization.
//
// TODO: can be removed once this ticket is fixed:
//       https://github.com/gogo/protobuf/issues/411#issuecomment-393856837
func (id *ID) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, b []byte) error {
	return id.UnmarshalJSON(b)
}

func marshalBytes(dst []byte, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		return 0, fmt.Errorf("buffer is too short")
	}
	return copy(dst, src), nil
}

var digits = [...]byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
	'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
	'U', 'V', 'W', 'X', 'Y', 'Z', '_', 'a', 'b', 'c',
	'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w',
	'x', 'y', 'z', '~'}

func encode(s *[11]byte, n uint64) {
	s[10], n = digits[n&0x3f], n>>6
	s[9], n = digits[n&0x3f], n>>6
	s[8], n = digits[n&0x3f], n>>6
	s[7], n = digits[n&0x3f], n>>6
	s[6], n = digits[n&0x3f], n>>6
	s[5], n = digits[n&0x3f], n>>6
	s[4], n = digits[n&0x3f], n>>6
	s[3], n = digits[n&0x3f], n>>6
	s[2], n = digits[n&0x3f], n>>6
	s[1], n = digits[n&0x3f], n>>6
	s[0] = digits[n&0x3f]
}
