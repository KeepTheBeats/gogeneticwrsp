package experimenttools

type cpuProcessor struct {
	name         string
	logicalCores float64
	baseClock    float64 // unit GHz
}

var cpuToChoose []cpuProcessor = append(amdEpyc7003Series, amdEpyc7002Series...)

// data from https://www.amd.com/en/processors/epyc-7003-series
var amdEpyc7003Series []cpuProcessor = []cpuProcessor{
	cpuProcessor{
		name:         "7773X",
		logicalCores: 128,
		baseClock:    2.2,
	},
	cpuProcessor{
		name:         "7763",
		logicalCores: 128,
		baseClock:    2.45,
	},
	cpuProcessor{
		name:         "7713P",
		logicalCores: 128,
		baseClock:    2.0,
	},
	cpuProcessor{
		name:         "7713",
		logicalCores: 128,
		baseClock:    2.0,
	},
	cpuProcessor{
		name:         "7663",
		logicalCores: 112,
		baseClock:    2.0,
	},
	cpuProcessor{
		name:         "7643",
		logicalCores: 96,
		baseClock:    2.3,
	},
	cpuProcessor{
		name:         "7573X",
		logicalCores: 64,
		baseClock:    2.8,
	},
	cpuProcessor{
		name:         "75F3",
		logicalCores: 64,
		baseClock:    2.95,
	},
	cpuProcessor{
		name:         "7543P",
		logicalCores: 64,
		baseClock:    2.8,
	},
	cpuProcessor{
		name:         "7543",
		logicalCores: 64,
		baseClock:    2.8,
	},
	cpuProcessor{
		name:         "7513",
		logicalCores: 64,
		baseClock:    2.6,
	},
	cpuProcessor{
		name:         "7473X",
		logicalCores: 48,
		baseClock:    2.8,
	},
	cpuProcessor{
		name:         "7453",
		logicalCores: 56,
		baseClock:    2.75,
	},
	cpuProcessor{
		name:         "74F3",
		logicalCores: 48,
		baseClock:    3.2,
	},
	cpuProcessor{
		name:         "7443P",
		logicalCores: 48,
		baseClock:    2.85,
	},
	cpuProcessor{
		name:         "7443",
		logicalCores: 48,
		baseClock:    2.85,
	},
	cpuProcessor{
		name:         "7413",
		logicalCores: 48,
		baseClock:    2.65,
	},
	cpuProcessor{
		name:         "73F3",
		logicalCores: 32,
		baseClock:    3.5,
	},
	cpuProcessor{
		name:         "7373X",
		logicalCores: 32,
		baseClock:    3.05,
	},
	cpuProcessor{
		name:         "7343",
		logicalCores: 32,
		baseClock:    3.2,
	},
	cpuProcessor{
		name:         "7313P",
		logicalCores: 32,
		baseClock:    3.0,
	},
	cpuProcessor{
		name:         "7313",
		logicalCores: 32,
		baseClock:    3.0,
	},
	cpuProcessor{
		name:         "72F3",
		logicalCores: 16,
		baseClock:    3.7,
	},
}

// data from https://www.amd.com/en/processors/epyc-7002-series
var amdEpyc7002Series []cpuProcessor = []cpuProcessor{
	cpuProcessor{
		name:         "7F72",
		logicalCores: 48,
		baseClock:    3.2,
	},
	cpuProcessor{
		name:         "7F52",
		logicalCores: 32,
		baseClock:    3.5,
	},
	cpuProcessor{
		name:         "7F32",
		logicalCores: 16,
		baseClock:    3.7,
	},
	cpuProcessor{
		name:         "7H12",
		logicalCores: 128,
		baseClock:    2.6,
	},
	cpuProcessor{
		name:         "7742",
		logicalCores: 128,
		baseClock:    2.25,
	},
	cpuProcessor{
		name:         "7702",
		logicalCores: 128,
		baseClock:    2.0,
	},
	cpuProcessor{
		name:         "7702P",
		logicalCores: 128,
		baseClock:    2.0,
	},
	cpuProcessor{
		name:         "7662",
		logicalCores: 128,
		baseClock:    2.0,
	},
	cpuProcessor{
		name:         "7642",
		logicalCores: 96,
		baseClock:    2.3,
	},
	cpuProcessor{
		name:         "7552",
		logicalCores: 96,
		baseClock:    2.2,
	},
	cpuProcessor{
		name:         "7542",
		logicalCores: 64,
		baseClock:    2.9,
	},
	cpuProcessor{
		name:         "7532",
		logicalCores: 64,
		baseClock:    2.4,
	},
	cpuProcessor{
		name:         "7502",
		logicalCores: 64,
		baseClock:    2.5,
	},
	cpuProcessor{
		name:         "7502P",
		logicalCores: 64,
		baseClock:    2.5,
	},
	cpuProcessor{
		name:         "7452",
		logicalCores: 64,
		baseClock:    2.35,
	},
	cpuProcessor{
		name:         "7402",
		logicalCores: 48,
		baseClock:    2.8,
	},
	cpuProcessor{
		name:         "7402P",
		logicalCores: 48,
		baseClock:    2.8,
	},
	cpuProcessor{
		name:         "7352",
		logicalCores: 48,
		baseClock:    2.3,
	},
	cpuProcessor{
		name:         "7302",
		logicalCores: 32,
		baseClock:    3.0,
	},
	cpuProcessor{
		name:         "7302P",
		logicalCores: 32,
		baseClock:    3.0,
	},
	cpuProcessor{
		name:         "7282",
		logicalCores: 32,
		baseClock:    2.8,
	},
	cpuProcessor{
		name:         "7272",
		logicalCores: 24,
		baseClock:    2.9,
	},
	cpuProcessor{
		name:         "7262",
		logicalCores: 16,
		baseClock:    3.2,
	},
	cpuProcessor{
		name:         "7252",
		logicalCores: 16,
		baseClock:    3.1,
	},
	cpuProcessor{
		name:         "7232P",
		logicalCores: 16,
		baseClock:    3.1,
	},
}
