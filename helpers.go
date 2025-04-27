package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
)

// Parses a string to d amount of decimals to float
func ParseFloatToXDecimals(n string, d int) (float64, error) {
	if d < 0 || d > 15 {
		return 0, fmt.Errorf("number of decimals must be non-negative and smaller than 15")
	}
	if n == "" {
		return 0, fmt.Errorf("string is empty")
	}
	if len(n) > 15 {
		return 0, fmt.Errorf("string too long")
	}

	v, err := strconv.ParseFloat(n, 64)
	log.Printf("Parsed to float: %f", v)
	if err != nil {
		return 0, fmt.Errorf("cannot parse float: %v", err)
	}

	decimals := math.Pow(10, float64(d))
	rounded := math.Round(v*decimals) / decimals 

	return rounded, nil
}

// Valid Coordinates
func IsValidLatitude(f float64) (bool, error) {
	if f > 90 || f < -90 {
		return false, nil
	}

	return true, nil
}

func IsValidLongitude(f float64) (bool, error) {
	if f > 180 || f < -180 {
		return false, nil
	}

	return true, nil
}

// Checks for negative nubmers
func IsNegative(f float64) (bool, error) {
	if f >= 0 {
		return false, nil
	}

	return true, nil
}

// Truncate a float32 to 2 decimals
func TruncateFloatTo2Decimals(f float64) float64 {
	return float64(int(f*100))/100
}
