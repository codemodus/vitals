package vitals

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
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

type MemStats struct {
	Allocs      uint64
	TotalAllocs uint64
	Sys         uint64
	Mallocs     uint64
	Frees       uint64
}

func NewMemStats(s *runtime.MemStats) *MemStats {
	return &MemStats{
		Allocs:      s.Alloc,
		TotalAllocs: s.TotalAlloc,
		Sys:         s.Sys,
		Mallocs:     s.Mallocs,
		Frees:       s.Frees,
	}
}

func (s *MemStats) String() string {
	return fmt.Sprintf(
		"CurAlloc(kB): %d, FromSys(kB): %d, CurMalloc: %d\n",
		s.Allocs/1000,
		s.Sys/1000,
		(s.Mallocs - s.Frees),
	)
}

func MemoryStats() *MemStats {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	return NewMemStats(ms)
}

func MonitorMemoryStats(cycle time.Duration) chan *MemStats {
	if cycle == 0 {
		return nil
	}

	c := make(chan *MemStats)

	go func() {
		memStats := &runtime.MemStats{}
		for {
			runtime.ReadMemStats(memStats)
			c <- NewMemStats(memStats)

			time.Sleep(cycle)
		}
	}()

	return c
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
