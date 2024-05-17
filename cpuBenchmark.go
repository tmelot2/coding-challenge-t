package main

import (
	"fmt"
	"github.com/dterei/gotsc"
	"golang.org/x/text/language"
    "golang.org/x/text/message"
	"strconv"
	"sync"
)

// Type-ify magic strings
type BenchmarkType string
const (
	BenchmarkTypeEnvFile        BenchmarkType = "envFile"
	BenchmarkTypeInputArgs      BenchmarkType = "inputArgs"
	BenchmarkTypeParseCsv       BenchmarkType = "parseCsv"
	BenchmarkTypeCalcQueryStats BenchmarkType = "calcQueryStats"
)

/*
   CpuBenchmark keeps track of CPU cycles* we want to measure in the query tool.

   It uses the gotsc measuring instrument to count cycles. Cycle measurement is
   done with assembly RDTSC/P instructions (and other assembly around the RDTRC
   to make sure the reading is accurate).

   * TODO: This doesn't measure *actual* cycles (yet), because modern RDTSC calls
   use invariant TSC, which does not scale up or down as CPUs adjust frequency. To
   get a more accurate reading, we need to get a calibration of a known timer,
   which is not yet implemented.
*/
type CpuBenchmark struct {
	tscOverhead	uint64

	// Tracks cycle counts for loading db configs from the .env file.
	readEnvFile         []uint64
	readEnvFileMutex    sync.Mutex

	// Tracks cycle counts for reading input args.
	readInputArgs       []uint64
	readInputArgsMutex  sync.Mutex

	// Tracks cycle counts for reading & parsing CSV lines
	parseCsv            []uint64
	parseCsvMutex       sync.Mutex

	// Tracks cycle counts for calculating all the query stats
	calcQueryStats      []uint64
	calcQueryStatsMutex sync.Mutex
}

func NewCpuBenchmark(capacity int) *CpuBenchmark {
	// Capacity must be positive
	if capacity < 0 {
		capacity = 0
	}

	readEnvFile    := make([]uint64, 0, capacity)
	readInputArgs  := make([]uint64, 0, capacity)
	parseCsv       := make([]uint64, 0, capacity)
	calcQueryStats := make([]uint64, 0, capacity)

	return &CpuBenchmark{
		tscOverhead:  gotsc.TSCOverhead(),

		readEnvFile:         readEnvFile,
		readEnvFileMutex:    sync.Mutex{},

		readInputArgs:       readInputArgs,
		readInputArgsMutex:  sync.Mutex{},

		parseCsv:            parseCsv,
		parseCsvMutex:       sync.Mutex{},

		calcQueryStats:      calcQueryStats,
		calcQueryStatsMutex: sync.Mutex{},
	}
}

// Adds cycle count to the list for the given type.
// NOTE: This will subtract TSC overhead for you!
func (b *CpuBenchmark) Add(bt BenchmarkType, cycles uint64) {
	count := cycles - b.tscOverhead

	switch bt {
	// Load db env file
	case BenchmarkTypeEnvFile:
		b.readEnvFileMutex.Lock()
		b.readEnvFile = append(b.readEnvFile, count)
		b.readEnvFileMutex.Unlock()
	// Read input args
	case BenchmarkTypeInputArgs:
		b.readInputArgsMutex.Lock()
		b.readInputArgs = append(b.readInputArgs, count)
		b.readInputArgsMutex.Unlock()
	// Read & parse CSV lines
	case BenchmarkTypeParseCsv:
		b.parseCsvMutex.Lock()
		b.parseCsv = append(b.parseCsv, count)
		b.parseCsvMutex.Unlock()
	// Calculate query stats
	case BenchmarkTypeCalcQueryStats:
		b.calcQueryStatsMutex.Lock()
		b.calcQueryStats = append(b.calcQueryStats, count)
		b.calcQueryStatsMutex.Unlock()
	}
}

func (b *CpuBenchmark) Print() {
	// Compute benchmark totals & grand total
	var totalCycles uint64

	// Env file
	var totalEnvFileCycles uint64
	for _,q := range b.readEnvFile {
		totalEnvFileCycles += q
	}
	totalCycles += totalEnvFileCycles

	// Read input args
	var totalReadInputArgCycles uint64
	for _,q := range b.readInputArgs {
		totalReadInputArgCycles += q
	}
	totalCycles += totalReadInputArgCycles

	// Read & parse CSV file
	var totalParseCsvCycles uint64
	for _,q := range b.parseCsv {
		totalParseCsvCycles += q
	}
	totalCycles += totalParseCsvCycles

	// Calculate query stats
	var totalCalcQueryStatsCycles uint64
	for _,q := range b.calcQueryStats {
		totalCalcQueryStatsCycles += q
	}
	totalCycles += totalCalcQueryStatsCycles

	fmt.Println("[CPU cycle totals]")

	// We should have > 0 measurements, if not, this arch isn't supported & the library just
	// returned 0 for cycle counts.
	if totalCycles == 0 {
		fmt.Println("CPU architecture not supported, only x86 works currently")
		return
	}

	// Calculate percents
	loadEnvPercent        := 100 * float64(totalEnvFileCycles)        / float64(totalCycles)
	readInputArgPercent   := 100 * float64(totalReadInputArgCycles)   / float64(totalCycles)
	parseCsvPercent       := 100 * float64(totalParseCsvCycles)       / float64(totalCycles)
	calcQueryStatsPercent := 100 * float64(totalCalcQueryStatsCycles) / float64(totalCycles)

	// Print with separators because cycle counts are large numbers
	width := 10
	p := message.NewPrinter(language.English)
    p.Printf( "   Load env file:  %*d  | %s%%\n", width, totalEnvFileCycles,        alignFloatAsStr(loadEnvPercent))
    p.Printf( " Read input args:  %*d  | %s%%\n", width, totalReadInputArgCycles,   alignFloatAsStr(readInputArgPercent))
    p.Printf( " Parse CSV lines:  %*d  | %s%%\n", width, totalParseCsvCycles,       alignFloatAsStr(parseCsvPercent))
    p.Printf( "Calc query stats:  %*d  | %s%%\n", width, totalCalcQueryStatsCycles, alignFloatAsStr(calcQueryStatsPercent))
    p.Println(" =================================================")
    p.Printf( "    Total cycles:  %*d\n\n", width, totalCycles)

    p.Printf("* This is not quite fully accurate cycle measurement (see TODO at top of cpuBenchmark.go).")
    p.Printf("  BUT, it does give a window into how long parts of the app are taking.")
}

func alignFloatAsStr(n float64) string {
	str := strconv.FormatFloat(n, 'f', 2, 64)
	if n < 10 {
		str = " " + str
	}
	return str
}
