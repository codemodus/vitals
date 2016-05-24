package vitals

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
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
				"CurAlloc(kB): %d, FromSys(kB): %d, CurMalloc: %d\n",
				memStats.Alloc/1000,
				memStats.Sys/1000,
				(memStats.Mallocs - memStats.Frees),
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

func SetupPIDFile() (deferFunc func(), err error) {
	pid := strconv.Itoa(os.Getpid())
	sn := path.Base(os.Args[0])
	dn := path.Join("/tmp", "."+sn+"-pid")
	fx := path.Join(dn, sn+".pid")

	if err = os.Mkdir(dn, 0700); os.IsNotExist(err) {
		return nil, err
	}

	f, err := os.Create(fx)
	if err != nil {
		return nil, err
	}

	if err = f.Truncate(0); err != nil {
		return nil, err
	}

	if _, err = f.WriteString(pid); err != nil {
		return nil, err
	}

	if err = f.Close(); err != nil {
		return nil, err
	}

	fn := func() {
		_ = os.RemoveAll(dn)
	}

	return fn, nil
}
