package analyse

import (
	"bufio"
	"fmt"
	"log/slog"
	"nganalysis/util"
	"os"
	"strconv"
	"strings"
)

func AnalyseLogs(files []*os.File, results *os.File, logger *slog.Logger) {
	var err error
	times := make([]float64, len(files))
	ramGo := make([]float64, len(files))
	ramLinux := make([]float64, len(files))

	util.LogInfo(logger, "Start analysis.")

	for i, file := range files {
		util.LogInfo(logger, fmt.Sprintf("analyse file (%d) %s", i+1, file.Name()))
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Training of") {
				elems := strings.Split(line, " ")
				times[i], err = strconv.ParseFloat(elems[7], 64)
				util.Panic(err, logger, fmt.Sprintf("could not parse float from file %s", file.Name()))
			}

			if strings.Contains(line, "\"total virtual") {
				elems := strings.Split(line, " ")
				ramGo[i], err = strconv.ParseFloat(elems[6], 64)
				util.Panic(err, logger, fmt.Sprintf("could not parse float from file %s", file.Name()))
			}

			if strings.Contains(line, "VmHWM") {
				elems := strings.Split(line, " ")
				ramLinux[i], err = strconv.ParseFloat(elems[8], 64)
				util.Panic(err, logger, fmt.Sprintf("could not parse float from file %s", file.Name()))
			}
		}
	}
	util.LogInfo(logger, "Compute final results.")

	timeMean, err := util.GetArithmeticMean(times, len(times))
	util.Panic(err, logger, "could not compute time mean")

	timeStdD, err := util.GetStdDeviation(times, times, len(times))
	util.Panic(err, logger, "could not compute std deviation")

	timeMax, err := util.Max(times...)
	util.Panic(err, logger, "could not compute max time")

	ramGoMean, err := util.GetArithmeticMean(ramGo, len(ramGo))
	util.Panic(err, logger, "could not compute mean ram go")

	ramGoMax, err := util.Max(ramGo...)
	util.Panic(err, logger, "could not comput max ram go")

	ramLinuxMean, err := util.GetArithmeticMean(ramLinux, len(ramLinux))
	util.Panic(err, logger, "could not compute mean ram linux")

	ramLinuxMax, err := util.Max(ramLinux...)
	util.Panic(err, logger, "could not comput max ram linux")

	util.LogInfo(logger, "Write final results to file.")

	results.WriteString(fmt.Sprintf("time Max:\t\t%.17f\n", timeMax))
	results.WriteString(fmt.Sprintf("time Mean:\t\t%.17f\n", timeMean))
	results.WriteString(fmt.Sprintf("time stdD:\t\t%.17f\n", timeStdD))

	results.WriteString(fmt.Sprintf("RAM (Go) Mean:\t\t%.17f\n", ramGoMean))
	results.WriteString(fmt.Sprintf("RAM (Go) Max:\t\t%.17f\n", ramGoMax))

	results.WriteString(fmt.Sprintf("RAM (Linux) Mean:\t%.17f\n", ramLinuxMean))
	results.WriteString(fmt.Sprintf("RAM (Linux) Max:\t%.17f\n", ramLinuxMax))
}
