package uuid

import (
	"crypto/rand"
	"fmt"
)

// The UUID reserved variants.
const (
	ReservedRFC4122 byte = 0x40
)

// A UUID representation
type UUID [16]byte

// Generate a random UUID.
func NewV4() (u *UUID, err error) {
	u = new(UUID)
	// Set all bits to randomly (or pseudo-randomly) chosen values.
	_, err = rand.Read(u[:])
	if err != nil {
		return
	}
	u.setVariant(ReservedRFC4122)
	u.setVersion(4)
	return
}

// Set the two most significant bits (bits 6 and 7) of the
// clock_seq_hi_and_reserved to zero and one, respectively.
func (u *UUID) setVariant(v byte) {
	switch v {
	case ReservedRFC4122:
		u[8] = (u[8] | ReservedRFC4122) & 0x7F
	}
}

// Set the four most significant bits (bits 12 through 15) of the
// time_hi_and_version field to the 4-bit version number.
func (u *UUID) setVersion(v byte) {
	u[6] = (u[6] & 0xF) | (v << 4)
}

// Returns unparsed version of the generated UUID sequence.
func (u *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}
