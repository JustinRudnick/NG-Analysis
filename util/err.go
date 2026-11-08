package util

import (
	"errors"
	"math"
)

func CircleWant(prototype []float64, center []float64) (p []float64, err error) {
	if len(prototype) != 2 {
		return nil, errors.New("wrong dimension, must have 2 elements")
	}

	prototypeCopy := make([]float64, len(prototype))
	for i := range prototypeCopy {
		prototypeCopy[i] = 2 * (prototype[i] - center[i]) //[-1, 1]
	}

	dist := EucDist(prototypeCopy, []float64{0, 0}) //distance of prototype to center (0,0)
	factor := 1.0 / dist                            //clean up factor - since it is a unit circle
	// fmt.Printf("EuDist: %.10f factor: %.10f\n", dist, factor)
	p = make([]float64, 2)
	p[0] = prototypeCopy[0]*factor/2.0 + center[0] //[0, 1]
	p[1] = prototypeCopy[1]*factor/2.0 + center[1] //[0, 1]

	return p, nil
}

// for EucDist(A, B, C) returns d(A, B) + d(B, C)
//
// for EucDist(A) it returns the vector's magnitude
//
// returns -1 if vectors have different dimensions
func EucDist(vecs ...[]float64) float64 {
	if len(vecs) == 0 {
		return 0
	}
	var lastpoint []float64
	if len(vecs) != 1 {
		lastpoint = vecs[0]
	} else {
		lastpoint = make([]float64, len(vecs[0]))
	}

	arr := make([]float64, len(vecs[0]))
	var dist float64

	for _, vec := range vecs {
		if len(vec) != len(lastpoint) {
			return -1
		}
		for i := range vec {
			arr[i] = vec[i] - lastpoint[i]
		}

		dist = Sum(arr, func(value float64, idx int) float64 { return value * value })
		dist = math.Sqrt(dist)

		lastpoint = vec
	}

	return dist
}
