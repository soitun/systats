package systats

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dhamith93/systats/internal/strops"
)

type CPU struct {
	LoadAvg   int
	CoreAvg   []int
	Model     string
	NoOfCores int
	Freq      string
	Cache     string
	Time      int64
}

func getCPU(systats *SyStats, milliseconds int) (CPU, error) {
	output := CPU{}
	statStr1, err := readFile(systats.StatFilePath)
	if err != nil {
		return output, err
	}
	// to calculate the cpu usage the /proc/stat has to be read some time apart
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
	statStr2, err := readFile(systats.StatFilePath)
	if err != nil {
		return output, err
	}

	processStatFileContents(&output, &statStr1, &statStr2)

	cpuinfoStr, err := readFile(systats.CPUinfoFilePath)
	if err != nil {
		return output, err
	}

	processCPUInfoFileContent(&output, &cpuinfoStr)

	return output, nil
}

func processCPUInfoFileContent(output *CPU, content *string) {
	split := strings.Split(*content, "\n")

	for _, line := range split {
		lineArr := strings.Fields(line)

		if len(lineArr) == 0 {
			continue
		}

		if len(lineArr) > 3 && (lineArr[0]+lineArr[1] == "modelname") {
			name := ""
			for i := 3; i < len(lineArr); i++ {
				name += lineArr[i] + " "
			}
			output.Model = strings.TrimSpace(name)
			continue
		}

		if len(lineArr) > 3 && (lineArr[0]+lineArr[1] == "cpucores") {
			output.NoOfCores, _ = strconv.Atoi(strings.TrimSpace(lineArr[3]))
			continue
		}

		if len(lineArr) > 3 && (lineArr[0]+lineArr[1] == "cpuMHz") {
			output.Freq = strings.TrimSpace(lineArr[3]) + " MHz"
			continue
		}

		if len(lineArr) > 3 && (lineArr[0]+lineArr[1] == "cachesize") {
			name := ""
			for i := 3; i < len(lineArr); i++ {
				name += lineArr[i] + " "
			}
			output.Cache = strings.TrimSpace(name)
			continue
		}
	}

	output.Time = time.Now().Unix()
}

func processStatFileContents(output *CPU, statStr1 *string, statStr2 *string) {
	statArr1 := processStatFile(statStr1)
	statArr2 := processStatFile(statStr2)

	for i := range statArr1 {
		// user + system, and user+system+idle times
		a1 := statArr1[i][0] + statArr1[i][1]
		a2 := statArr1[i][0] + statArr1[i][1] + statArr1[i][2]
		usage := 100 * (statArr2[i][0] + statArr2[i][1] - a1) / (statArr2[i][0] + statArr2[i][1] + statArr2[i][2] - a2)

		if i == 0 {
			output.LoadAvg = int(usage)
			continue
		}
		output.CoreAvg = append(output.CoreAvg, int(usage))
	}
}

func processStatFile(content *string) [][]uint64 {
	statSplit := strings.Split(*content, "\n")
	output := [][]uint64{}
	r, _ := regexp.Compile("^cpu")

	for _, line := range statSplit {
		lineArr := strings.Fields(line)
		if len(lineArr) > 0 && r.MatchString(lineArr[0]) {
			output = append(output, []uint64{strops.ToUint64(lineArr[1]), strops.ToUint64(lineArr[3]), strops.ToUint64(lineArr[4])})
		}
	}

	return output
}