package base

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// Rat is identical to big.Rat, except it can implements the
// json.Marshaler interface so that it can also accept integers.
type Rat big.Rat

// MarshalJSON implements the json.Marshaler interface. If the rational is an
// integer and it fits on a IEEE 754 number, it will be marshaled as a JSON
// number. Otherwise it will be marshaled as a string.
func (r *Rat) MarshalJSON() ([]byte, error) {
	if (*big.Rat)(r).IsInt() && (*big.Rat)(r).Num().BitLen() <= 53 {
		return (*big.Rat)(r).MarshalText()
	}
	return []byte(fmt.Sprintf("%q", (*big.Rat)(r).RatString())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. If the rational is
// a number, it will be parsed as one. Otherwise, it will use
// big.Rat.UnmarshalText.
func (r *Rat) UnmarshalJSON(data []byte) error {
	var i big.Int
	err := i.UnmarshalJSON(data)
	if err == nil {
		(*big.Rat)(r).SetInt(&i)
		return nil
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("rat: %w", err)
	}
	return (*big.Rat)(r).UnmarshalText([]byte(s))
}

// ParseRational returns a rational that's within 1e-6 of the floating-point
// value that has been serialized as a string.
func ParseRational(str string) (*big.Rat, error) {
	floatVal, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, err
	}

	return FloatToRational(floatVal), nil
}

// FloatToRational returns a rational that's within 1e-6 of the floating-point
// value.
func FloatToRational(floatVal float64) *big.Rat {
	tolerance := 1e-6

	var chosenDenominator int64 = 1024
	for _, denominator := range []int64{1, 100, 1000, 720, 5040, 40320, 3628800} {
		scaledVal := floatVal * float64(denominator)
		if math.Abs(scaledVal-math.Round(scaledVal)) <= tolerance {
			chosenDenominator = denominator
			break
		}
	}

	return big.NewRat(
		int64(math.Round(floatVal*float64(chosenDenominator))),
		chosenDenominator,
	)
}

// RationalToFloat returns the closest float value to the given big.Rat.
func RationalToFloat(val *big.Rat) float64 {
	floatVal, _ := val.Float64()
	return floatVal
}

// RationalDiv implements division between two rationals.
func RationalDiv(num, denom *big.Rat) *big.Rat {
	val := new(big.Rat).SetFrac(denom.Denom(), denom.Num())
	return val.Mul(val, num)
}
