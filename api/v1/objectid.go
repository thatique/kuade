package v1

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

var ErrInValidObjectID = errors.New("the provided id representation is not a valid ObjectID")

// ObjectID is the BSON ObjectID type.
type ObjectID [12]byte

// NilObjectID is the zero value for ObjectID.
var NilObjectID ObjectID

var objectIDCounter = readRandomUint32()
var processUnique = processUniqueBytes()

// NewObjectID generates a new ObjectID.
func NewObjectID() ObjectID {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[0:4], uint32(time.Now().Unix()))

	copy(b[4:9], processUnique[:])
	putUint24(b[9:12], atomic.AddUint32(&objectIDCounter, 1))

	return b
}

func NewObjectIdWithTime(t time.Time) ObjectID {
	var b [12]byte
	binary.BigEndian.PutUint32(b[0:4], uint32(t.Unix()))

	copy(b[4:9], processUnique[:])
	putUint24(b[9:12], atomic.AddUint32(&objectIDCounter, 1))

	return b
}

// ObjectIDFromHex creates a new ObjectID from a hex string. It returns an error if the hex string is not a
// valid ObjectID.
func ObjectIDFromHex(s string) (ObjectID, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return NilObjectID, err
	}

	if len(b) != 12 {
		return NilObjectID, ErrInValidObjectID
	}

	var oid [12]byte
	copy(oid[:], b[:])

	return oid, nil
}

func (oid ObjectID) Hex() string {
	return hex.EncodeToString(oid[:])
}

func (oid ObjectID) String() string {
	return fmt.Sprintf("ObjectID(%q)", oid.Hex())
}

// IsZero returns true if id is the empty ObjectID.
func (oid ObjectID) IsZero() bool {
	return bytes.Equal(oid[:], NilObjectID[:])
}

// Size returns the size of this datum in protobuf. It is always 16 bytes.
func (oid ObjectID) Size() int {
	return 12
}

func (oid ObjectID) Time() time.Time {
	secs := int64(binary.BigEndian.Uint32(oid[0:4]))
	return time.Unix(secs, 0)
}

func (oid ObjectID) MarshalJSON() ([]byte, error) {
	var b [12]byte
	oid.MarshalTo(b[:])
	// the hex representation size is 24 + 2 for double quote
	s := make([]byte, 26)
	hex.Encode(s[1:25], b[:])
	// put the quote
	s[0], s[25] = '"', '"'
	return s, nil
}

func (oid *ObjectID) UnmarshalJSON(data []byte) error {
	if len(data) != 26 {
		return ErrInValidObjectID
	}
	if data[0] != '"' || data[25] != '"' {
		return ErrInValidObjectID
	}
	dst := make([]byte, 12)
	n, err := hex.Decode(dst, data[1:25])
	if err != nil {
		return err
	}
	if n != 12 {
		return ErrInValidObjectID
	}
	copy(oid[:], dst)
	return nil
}

// MarshalTo converts trace ID into a binary representation. Called by protobuf serialization.
func (oid ObjectID) MarshalTo(data []byte) (n int, err error) {
	var b [12]byte
	copy(b[:], oid[:])
	return marshalBytes(data, b[:])
}

// Unmarshal inflates this trace ID from binary representation. Called by protobuf serialization.
func (oid *ObjectID) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("buffer is too short")
	}
	copy(oid[:], data)
	return nil
}

func processUniqueBytes() [5]byte {
	var b [5]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return b
}

func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func putUint24(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func marshalBytes(dst []byte, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		return 0, fmt.Errorf("buffer is too short")
	}
	return copy(dst, src), nil
}
