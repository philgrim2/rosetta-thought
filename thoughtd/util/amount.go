// Copyright (c) 2013, 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util

import (
	"errors"
	"math"
	"strconv"
)

// AmountUnit describes a method of converting an Amount to something
// other than the base unit of a thought.  The value of the AmountUnit
// is the exponent component of the decadic multiple to convert from
// an amount in thought to an amount counted in units.
type AmountUnit int

// These constants define various units used when describing a thought
// monetary amount.
const (
	AmountMegaTHT  AmountUnit = 6
	AmountKiloTHT  AmountUnit = 3
	AmountTHT      AmountUnit = 0
	AmountMilliTHT AmountUnit = -3
	AmountMicroTHT AmountUnit = -6
	AmountNotion  AmountUnit = -8
	NotionPerThought = 1e8
)

// String returns the unit as a string.  For recognized units, the SI
// prefix is used, or "Notion" for the base unit.  For all unrecognized
// units, "1eN THT" is returned, where N is the AmountUnit.
func (u AmountUnit) String() string {
	switch u {
	case AmountMegaTHT:
		return "MTHT"
	case AmountKiloTHT:
		return "kTHT"
	case AmountTHT:
		return "THT"
	case AmountMilliTHT:
		return "mTHT"
	case AmountMicroTHT:
		return "μTHT"
	case AmountNotion:
		return "Notion"
	default:
		return "1e" + strconv.FormatInt(int64(u), 10) + " THT"
	}
}

// Amount represents the base thought monetary unit (colloquially referred
// to as a `Notion').  A single Amount is equal to 1e-8 of a thought.
type Amount int64

// round converts a floating point number, which may or may not be representable
// as an integer, to the Amount integer type by rounding to the nearest integer.
// This is performed by adding or subtracting 0.5 depending on the sign, and
// relying on integer truncation to round the value to the nearest Amount.
func round(f float64) Amount {
	if f < 0 {
		return Amount(f - 0.5)
	}
	return Amount(f + 0.5)
}

// NewAmount creates an Amount from a floating point value representing
// some value in thought.  NewAmount errors if f is NaN or +-Infinity, but
// does not check that the amount is within the total amount of thought
// producible as f may not refer to an amount at a single moment in time.
//
// NewAmount is for specifically for converting THT to Notion.
// For creating a new Amount with an int64 value which denotes a quantity of Notion,
// do a simple type conversion from type int64 to Amount.
// See GoDoc for example: http://godoc.org/github.com/btcsuite/btcd/btcutil#example-Amount
func NewAmount(f float64) (Amount, error) {
	// The amount is only considered invalid if it cannot be represented
	// as an integer type.  This may happen if f is NaN or +-Infinity.
	switch {
	case math.IsNaN(f):
		fallthrough
	case math.IsInf(f, 1):
		fallthrough
	case math.IsInf(f, -1):
		return 0, errors.New("invalid thought amount")
	}

	return round(f * NotionPerThought), nil
}

// ToUnit converts a monetary amount counted in thought base units to a
// floating point value representing an amount of thought.
func (a Amount) ToUnit(u AmountUnit) float64 {
	return float64(a) / math.Pow10(int(u+8))
}

// ToTHT is the equivalent of calling ToUnit with AmountTHT.
func (a Amount) ToTHT() float64 {
	return a.ToUnit(AmountTHT)
}

// Format formats a monetary amount counted in thought base units as a
// string for a given unit.  The conversion will succeed for any unit,
// however, known units will be formated with an appended label describing
// the units with SI notation, or "Notion" for the base unit.
func (a Amount) Format(u AmountUnit) string {
	units := " " + u.String()
	return strconv.FormatFloat(a.ToUnit(u), 'f', -int(u+8), 64) + units
}

// String is the equivalent of calling Format with AmountTHT.
func (a Amount) String() string {
	return a.Format(AmountTHT)
}

// MulF64 multiplies an Amount by a floating point value.  While this is not
// an operation that must typically be done by a full node or wallet, it is
// useful for services that build on top of thought (for example, calculating
// a fee by multiplying by a percentage).
func (a Amount) MulF64(f float64) Amount {
	return round(float64(a) * f)
}
