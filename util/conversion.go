package util

import (
	"gonum.org/v1/gonum/mat"
)

func SliceToVecDense(slices ...[]float64) []*mat.VecDense {
	vecs := make([]*mat.VecDense, len(slices))
	for i, slice := range slices {
		vecs[i] = mat.NewVecDense(len(slice), slice)
	}
	return vecs
}

func VecDenseToSlice(vecs ...*mat.VecDense) [][]float64 {
	slice := make([][]float64, len(vecs))
	for i, vec := range vecs {
		slice[i] = vec.RawVector().Data
	}
	return slice
}
