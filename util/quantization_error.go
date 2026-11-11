package util

import (
	"errors"
	"math"
)

// calculates the quantiztaion error by the cost function (8) in the Martinetz Paper from 1993
// ("Neural-Gas" Network for Vector Quantization and its Application to Time-Series Prediction)
//
// Note that there are more samples, than prototypes
func GetQuantizationEMartinetz(prototype []float64, samples [][]float64, prototypes [][]float64, lambda float64) (res float64) {
	sampleC := len(samples)
	prototypeC := len(prototypes)

	distSq := func(p1, p2 []float64) float64 {
		return Sum[float64](p1, func(value float64, idx int) float64 { return math.Pow(p1[idx]-p2[idx], 2) })
	}

	//function to compute rank of passed prototype to the passed sample
	rank := func(sample, prototype []float64) int {
		dCurr := distSq(sample, prototype)
		rank := 0
		for _, p := range prototypes {
			if dCurr > distSq(sample, p) {
				rank++
			}
		}
		return rank
	}

	lastpart := func(sampleI, prototypeI int, lambda float64) float64 {
		return math.Exp(-float64(rank(samples[sampleI], prototype))/lambda) * distSq(samples[sampleI], prototypes[prototypeI])
	}

	_C := func(lambda float64, prototypeCount int) float64 {
		ranks := make([]int, prototypeCount)
		for i := range prototypeCount {
			ranks[i] = i
		}
		return SumAny[int, float64](ranks, func(value, idx int) float64 { return math.Exp(-float64(value) / lambda) })
	}

	factor := 0.5 / _C(lambda, prototypeC) / float64(sampleC)

	var sum float64 = 0
	for j := range sampleC {
		for i := range prototypeC {
			sum += lastpart(j, i, lambda)
		}
	}

	return factor * sum
}

/*
For M samples and N prototypes and let d denote the euclidian distance between two points.
Then this function calculates the quantization error as follows:

1/M * sum_1^M(min_1^N(d(sample, prototype)))
*/
func GetQuantizationE(prototype []float64, samples [][]float64, prototypes [][]float64, lambda float64) (res float64) {
	factor := float64(1) / float64(len(samples))

	minDistance := func(sample []float64) float64 {
		bestFittingPrototype := MinAny(
			func(proto, currMin []float64) bool {
				return EucDist(sample, proto) < EucDist(sample, currMin)
			}, prototypes...)
		return EucDist(bestFittingPrototype, sample)
	}

	quantizationErrorAbs := SumAny(samples, func(value []float64, idx int) float64 {
		minD := minDistance(value)
		return minD * minD
	})

	return factor * quantizationErrorAbs

}

// mean quantization error
func GetMQE(samples [][]float64, prototypes [][]float64, lambda float64) (res float64) {
	var sum float64 = 0
	for _, p := range prototypes {
		sum += GetQuantizationE(p, samples, prototypes, lambda)
	}
	return sum / float64(len(prototypes))
}

func GetMaxQE(samples [][]float64, prototypes [][]float64, lambda float64) (res float64) {
	var maxQE float64 = 0
	for _, p := range prototypes {
		maxQE = math.Max(maxQE, GetQuantizationE(p, samples, prototypes, lambda))
	}
	return maxQE
}

// get absolute error (AE)
//
// AE = abs(want - have)
func GetAE(want float64, have float64) (res float64) {
	return math.Abs(want - have)
}

// Get mean absolute error (MAE) of the first n elements
//
// MAE = sum(AE)/n
func GetMAE(want []float64, have []float64, n int) (res float64, err error) {
	return _GetMPowE(want, have, n, 1)
}

// Get mean square error (MSE) of the first n elements
//
// MSE = sum(AE * AE)/n
func GetMSE(want []float64, have []float64, n int) (res float64, err error) {
	return _GetMPowE(want, have, n, 2)
}

// Get root mean square error (MSE) of the first n elements
//
// MSE = sqrt(sum(AE * AE)/n)
func GetRMSE(want []float64, have []float64, n int) (res float64, err error) {
	if res, err = GetMSE(want, have, n); err != nil {
		return 0, err
	}
	return math.Sqrt(res), nil
}

// Get relative error (RE) of the first n elements
//
// RE = |have - want|/|want|
func GetRE(want float64, have float64) (res float64) {
	return math.Abs(want-have) / math.Abs(want)
}

// Get relative percentage  error (RPE) of the first n elements
//
// RPE = 100 * RE
func GetRPE(want float64, have float64) (res float64) {
	return 100 * GetRE(want, have)
}

// Get the maximum absolue error (MaxAE) of the first n elements
//
// MaxAE = max(GetAE...)
func GetMaxAE(want []float64, have []float64, n int) (res float64, err error) {
	return _GetMaxE(want, have, n, GetAE)
}

// Get the maximum relative error (MaxRE) of the first n elements
//
// MaxRE = max(GetRE...)
func GetMaxRE(want []float64, have []float64, n int) (res float64, err error) {
	return _GetMaxE(want, have, n, GetRE)
}

// Get mean absolute percentage error (MAPE) of the first n elements
//
// MAPE = 100 * MAE
func GetMAPE(want []float64, have []float64, n int) (res float64, err error) {
	if res, err = GetMAE(want, have, n); err != nil {
		return 0, err
	}
	return 100 * res, nil
}

func GetStdDeviation(want []float64, have []float64, n int) (res float64, err error) {
	variance, err := GetVariance(want, have, n)
	return math.Sqrt(variance), err
}

func GetVariance(want []float64, have []float64, n int) (res float64, err error) {
	if n <= 1 {
		return 0, errors.New("too small n")
	}

	if len(want) < n || len(have) < n {
		err = errors.New("buffer overflow error.")
		return 0, err
	}

	var arithmeticMean float64
	if arithmeticMean, err = GetArithmeticMean(want, n); err != nil {
		return 0, err
	}

	var sum float64 = 0
	for i := range n {
		sum += math.Pow(have[i]-arithmeticMean, 2)
	}

	return sum / (float64(n) - 1), nil
}

//########################## print functions #####################################################

func PrintSlot(want float64, have float64) {
	println(want, "\t", have, "\t", GetAE(want, have), "\t", GetRE(want, have))
}

func PrintSlots(want []float64, have []float64, n int) {

	println("WANT\t\t\tHAVE\t\t\tabs. ERROR\t\t\trel. ERROR")

	for i := range n {
		PrintSlot(want[i], have[i])
	}

	println()
	println("MAE:\t\t", _printValuetNoErr(want, have, n, GetMAE))
	println("MSE:\t\t", _printValuetNoErr(want, have, n, GetMSE))
	println("RMSE:\t\t", _printValuetNoErr(want, have, n, GetRMSE))
	println("Max AE:\t\t", _printValuetNoErr(want, have, n, GetMaxAE))
	println("Max RE:\t\t", _printValuetNoErr(want, have, n, GetMaxRE))
	println("std. deviation:\t", _printValuetNoErr(want, have, n, GetStdDeviation))
	println("variance:\t", _printValuetNoErr(want, have, n, GetVariance))
}

//########################## helper functions ####################################################

// Get mean pow error (MPowE) of the first n elements
//
// MPowE = sum(AE^pow)/n
func _GetMPowE(want []float64, have []float64, n int, pow float64) (res float64, err error) {

	if len(want) < n || len(have) < n {
		err = errors.New("buffer overflow error.")
		return 0, err
	}

	var sum float64 = 0
	for i := range n {
		sum += math.Pow(GetAE(want[i], have[i]), pow)
	}

	return sum / float64(n), nil
}

func _GetMaxE(want []float64, have []float64, n int, fnc func(want float64, have float64) (res float64)) (res float64, err error) {
	if len(want) < n || len(have) < n {
		err = errors.New("buffer overflow error.")
		return 0, err
	}

	var max float64 = 0
	for i := range n {
		max = math.Max(fnc(want[i], have[i]), max)
	}
	return max, nil
}

func GetArithmeticMean(slice []float64, n int) (res float64, err error) {
	if len(slice) < n {
		err = errors.New("buffer overflow error.")
		return 0, err
	}

	var sum float64 = 0
	for i := range n {
		sum += slice[i]
	}
	return sum / float64(n), nil

}

func _printValuetNoErr(want []float64, have []float64, n int, fnc func(want []float64, have []float64, n int) (res float64, err error)) (value float64) {
	var err error

	if value, err = fnc(want, have, n); err != nil {
		return math.NaN()
	}

	return value
}
