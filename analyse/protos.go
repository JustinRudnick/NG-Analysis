package analyse

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	plotting "nganalysis/Plotting"
	"nganalysis/input"
	"nganalysis/util"
	"os"

	"gonum.org/v1/gonum/mat"
)

// wants ...sample csvs corresponding to
func AnalyseProtos(sampleC int,
	headed bool,
	split, trim string,
	prototypeFiles []*os.File,
	sampleSetFilePaths []string,
	sampleFileType input.SampleInputType,
	CsvDimension int, // if sampleFileType is input.CSV_NO_HEADER this one is needed
	CsvDelim string, //delimiter of sample file
	results *os.File,
	logger *slog.Logger,
) {
	if len(prototypeFiles) != len(sampleSetFilePaths) {
		util.Panic(errors.New("amount of passed prototype files and amount of passed sample set files differ."), logger, "amount of passed prototype files and amount of passed sample set files differ.")
	}
	csvs, err := util.NewCSVFiles(headed, split, trim, prototypeFiles)
	util.Panic(err, logger, "could not create CSV objects")

	values := make([][]float64, 0)

	qes := make([]float64, 0)                      // buffer of all QEs over all csvs
	mQEexp := make([]float64, len(prototypeFiles)) //mean per incoming file

	var maxQE float64 = 0
	var meanQE float64 = 0

	var maxAE float64 = 0
	var meanAEs []float64 = make([]float64, len(prototypeFiles))

	for i, csv := range csvs {
		logger.Info(fmt.Sprintf("Entering csv (%d) %s", i+1, csv.File.Name()))

		// read in all prototype vectors
		prototypes := make([][]float64, 0)
		for csv.Scan() {
			value, err := csv.ReadLine()
			util.Panic(err, logger, "could not read from csv")

			prototypes = append(prototypes, value)
		}
		values = append(values, prototypes...)

		var sampleSet []*mat.VecDense
		switch sampleFileType {
		case input.PLY:
			sampleSet, err = input.PlyToSampleSet(sampleSetFilePaths[i])
		case input.CSV_NO_HEADER:
			sampleSet, err = input.CSVNoHeaderToSampleSet(sampleSetFilePaths[i], CsvDimension, logger)
			sampleSet = input.NormalizeVecDense(sampleSet)
		}
		util.Panic(err, logger, fmt.Sprintf("could not read sample set from: %s", sampleSetFilePaths[i]))

		plotting.Plot2D(sampleSet, "TestBunny", fmt.Sprintf("TestPlot%d", i))

		samples := util.VecDenseToSlice(sampleSet...)

		//
		lambda := 0.01
		maxQE, err = util.Max(maxQE, util.GetMaxQE(samples, prototypes, lambda))
		maxQE = math.Max(maxQE, util.GetMaxQE(samples, prototypes, lambda))
		mQEexp[i] = util.GetMQE(samples, prototypes, lambda)
		for _, value := range prototypes {
			qes = append(qes, util.GetQuantizationE(value, samples, prototypes, lambda))
		}

		isCloser := func(sample, currNearest *mat.VecDense) bool {
			return util.EucDist(sample.RawVector().Data) < util.EucDist(currNearest.RawVector().Data)
		}
		getNearestSample := func(prototype *mat.VecDense, samples []*mat.VecDense) *mat.VecDense {
			samplesCopy := make([]*mat.VecDense, len(samples))
			copy(samplesCopy, samples)
			for i, sample := range samplesCopy {
				samplesCopy[i].SubVec(sample, prototype)
			}
			return util.MinAny(isCloser, samplesCopy...)
		}

		AEsV := make([]*mat.VecDense, len(prototypes))
		for i, prototype := range util.SliceToVecDense(prototypes...) {
			AEsV[i] = getNearestSample(prototype, sampleSet)
		}
		AEs := util.ForAll(func(idx int, value *mat.VecDense) float64 { return util.EucDist(value.RawVector().Data) }, AEsV...)

		maxAERun, err := util.Max(AEs...)
		util.Panic(err, logger, "Could not find the maximum value of absolute error.")
		maxAE = math.Max(maxAE, maxAERun)
		meanAERun, err := util.GetArithmeticMean(AEs, len(AEs))
		util.Panic(err, logger, "Could not compute arithmetic mean of absolute errors.")
		meanAEs[i] = meanAERun
	}
	meanQE = util.SumElems(mQEexp...) / float64(len(csvs))
	stdDevMQE, err := util.GetStdDeviation(mQEexp, mQEexp, len(mQEexp))
	if len(prototypeFiles) == 1 && err != nil {
		logger.Warn("could not compute std deviation of mQE - too few prototype files (set to 0)")
	} else {
		util.Panic(err, logger, "could not compute std deviation of mQE") //cannot compute other results
	}

	//calculate rmse of qes
	for i := range qes {
		qes[i] *= qes[i]
	}
	mse := util.SumElems(qes...) / float64(len(qes))
	rmse := math.Sqrt(mse)

	util.LogInfo(logger, "Loop done")

	results.WriteString(fmt.Sprintf("max QE:\t\t%.17f\n", maxQE))
	results.WriteString(fmt.Sprintf("MQE:\t\t%.17f\n", meanQE))
	results.WriteString(fmt.Sprintf("stdDevMQE:\t%.17f\n", stdDevMQE))

	mae, err := util.GetArithmeticMean(meanAEs, len(meanAEs))
	util.Panic(err, logger, "could not extract MAE.")

	results.WriteString(fmt.Sprintf("max AE:\t\t%.17f\n", maxAE))
	results.WriteString(fmt.Sprintf("MAE:\t\t%.17f\n", mae))
	results.WriteString(fmt.Sprintf("RMSQE:\t\t%.17f\n", rmse))
}
