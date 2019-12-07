package base

import (
	"math"
	"math/big"
	"strconv"
)

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
