package main

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"nganalysis/analyse"
	"nganalysis/input"
	"nganalysis/util"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/mat"
)

// errors
// diagram

func main() {

	if util.In([]string{"-help", "h", "?", "-h", "help"}, os.Args[1]) {
		fmt.Printf("%s <protos | log>\n", os.Args[0])
		fmt.Printf("[log]: <dir> <regex file search> <result file>\n")
		fmt.Printf("[protos]: <dir> <regex file search> <result file> <sample count> <ply | csvnh>\n")
		fmt.Printf("[protos/ply]: <sample dir> <regex sample files>\n")
		fmt.Printf("[protos/csvnh]: <sample dir> <regex sample files> <sample data dimension> <split/delim>\n")
		return
	}
	Args_Option := 1 // "ply" | "csvnh"
	Args_Dir := Args_Option + 1
	Args_RegexProtos := Args_Dir + 1
	Args_ResultFile := Args_RegexProtos + 1
	// [protos]
	Args_ProtosSampleCount := Args_ResultFile + 1
	Args_ProtosSampleFileType := Args_ProtosSampleCount + 1
	// [protos/ply]
	Args_ProtosPlySampleDir := Args_ProtosSampleFileType + 1
	Args_ProtosPlySampleRegex := Args_ProtosPlySampleDir + 1
	// [protos/csvnh]
	Args_ProtosCsvnhSampleDir := Args_ProtosSampleFileType + 1
	Args_ProtosCsvnhSampleRegex := Args_ProtosCsvnhSampleDir + 1
	Args_ProtosCsvnhDimension := Args_ProtosCsvnhSampleRegex + 1
	Args_ProtosCsvnhDelim := Args_ProtosCsvnhDimension + 1

	var err error
	var logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	dir := os.Args[Args_Dir]
	regex := os.Args[Args_RegexProtos]
	results, err := os.OpenFile(os.Args[Args_ResultFile], os.O_CREATE|os.O_RDWR, 0644)
	util.Panic(err, logger, "open file failed")
	defer results.Close()

	results.Truncate(0)

	files, err := util.OpenFiles(dir, false, regex, os.O_RDONLY, 0644, logger)
	util.Panic(err, logger, "open files failed")
	defer util.CloseFiles(files)

	logger.Info(fmt.Sprintf("%d files match regular expression.", len(files)))

	if util.In([]string{"p", "-p", "protos"}, os.Args[Args_Option]) {
		headed := false
		split := ","
		trim := " "
		sampleC, err := strconv.Atoi(os.Args[Args_ProtosSampleCount])
		util.Panic(err, logger, "could not parse sample count.")

		fileType := os.Args[Args_ProtosSampleFileType]
		util.Panic(err, logger, "could not convert [sample file type] to integer.")

		switch strings.ToLower(fileType) {
		case "csvnh":
			sampleFilePaths, err := util.RegexMatchFilePaths(os.Args[Args_ProtosCsvnhSampleDir], false, os.Args[Args_ProtosCsvnhSampleRegex], logger)
			util.Panic(err, logger, "Could not resolve file paths.", "RegexMatchFilePaths failed:")

			csvDim, err := strconv.Atoi(os.Args[Args_ProtosCsvnhDimension])
			util.Panic(err, logger, fmt.Sprintf("Input %s must be integer.", os.Args[Args_ProtosCsvnhDimension]))

			analyse.AnalyseProtos(sampleC, headed, split, trim, files, sampleFilePaths, input.CSV_NO_HEADER, csvDim, os.Args[Args_ProtosCsvnhDelim], results, logger)

		case "ply":
			sampleFilePaths, err := util.RegexMatchFilePaths(os.Args[Args_ProtosPlySampleDir], false, os.Args[Args_ProtosPlySampleRegex], logger)
			util.Panic(err, logger, "Could not resolve file paths.", "RegexMatchFilePaths failed:")

			analyse.AnalyseProtos(sampleC, headed, split, trim, files, sampleFilePaths, input.PLY, 0, "\t", results, logger)

		default:
			msg := fmt.Sprintf("Option %s not supported", os.Args[Args_Option])
			util.Panic(errors.New(msg), logger, msg)
		}

	} else if util.In([]string{"l", "-l", "log"}, os.Args[Args_Option]) {
		analyse.AnalyseLogs(files, results, logger)
	}

}

func fillDataset(dataset []*mat.VecDense, RNG *rand.Rand) {
	for i := range len(dataset) {
		rng := RNG.Float64()
		dataset[i] = mat.NewVecDense(2, []float64{0.5*math.Sin(rng*2*math.Pi) + 0.5, 0.5*math.Cos(rng*2*math.Pi) + 0.5}) //circle
		// dataset[2*i] = mat.NewVecDense(2, []float64{rng, math.Cos(rng)})	// sin cos (1/2)
		// dataset[2*i+1] = mat.NewVecDense(2, []float64{rng, math.Sin(rng)}) // sin cos (2/2)
		// dataset[i] = mat.NewVecDense(2, []float64{0.5*rng + 0.2, 0.2*rand.Float64() + 0.4}) // rectangle area
	}
}
