package vitals

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	nullLog = log.New(ioutil.Discard, "", 0)
)

func StartCPUProfile(file string) (deferFunc func(), err error) {
	if file == "" {
		return func() {}, nil
	}

	f, err := os.Create(file)
	if err != nil {
		return func() {}, err
	}
	if err = pprof.StartCPUProfile(f); err != nil {
		return func() {}, err
	}

	fn := func() {
		pprof.StopCPUProfile()
	}
	return fn, nil
}

func LogMemStats(cycle time.Duration, logger *log.Logger) {
	if cycle == 0 {
		return
	}

	if logger == nil {
		logger = nullLog
	}

	go func() {
		memStats := runtime.MemStats{}
		for {
			runtime.ReadMemStats(&memStats)
			logger.Printf(
				"%d, %d, %d, %d\n",
				memStats.HeapSys,
				memStats.HeapAlloc,
				memStats.HeapIdle,
				memStats.HeapReleased,
			)
			time.Sleep(cycle)
		}
	}()
}

func WriteHeapProfile(file string) error {
	if file == "" {
		return nil
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if err = pprof.WriteHeapProfile(f); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}
	return nil
}
