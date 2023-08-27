package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	factor     = 5
	numRecords = 10000000
)

type Employment struct {
	Company          string
	Position         string
	StartDate        time.Time
	EndDate          time.Time
	Responsibilities []string
}

// new prometheus register
var (
	prometheusRegister = prometheus.NewRegistry()
	pusher             *push.Pusher
)

// metrics
var (
	metricStartTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "start_timestamp",
		Help:      "Start timestamp",
		Namespace: "util",
	})
	metricNumberOfCPU = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "number_of_cpu",
		Help:      "Number of CPU",
		Namespace: "util",
	})
	metricNumberOfGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "number_of_goroutines",
		Help:      "Number of goroutines",
		Namespace: "util",
	})
	metricChunkSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "chunk_size",
		Help:      "Chunk size",
		Namespace: "util",
	})
	metricActualNumberOfGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "actual_number_of_goroutines",
		Help:      "Actual number of goroutines",
		Namespace: "util",
	})
	metricNumberOfGC = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "number_of_gc",
		Help:      "Number of GC",
		Namespace: "util",
	})
	metricAllocatedMemory = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:      "allocated_memory",
		Help:      "Allocated memory",
		Namespace: "util",
		Buckets:   prometheus.LinearBuckets(0, 100000000, 10),
	})
)

func init() {
	prometheusRegister.MustRegister(metricStartTimestamp)
	prometheusRegister.MustRegister(metricNumberOfCPU)
	prometheusRegister.MustRegister(metricNumberOfGoroutines)
	prometheusRegister.MustRegister(metricChunkSize)
	prometheusRegister.MustRegister(metricActualNumberOfGoroutines)
	prometheusRegister.MustRegister(metricNumberOfGC)
	prometheusRegister.MustRegister(metricAllocatedMemory)

	pusher = push.New("http://localhost:9091", "util").Gatherer(prometheusRegister)
}

func main() {
	defer func() {
		if err := pusher.Add(); err != nil {
			log.Fatal(err)
		}
	}()

	var wg sync.WaitGroup

	metricStartTimestamp.SetToCurrentTime()

	recordChan := make(chan Employment, 100000)
	defer close(recordChan)

	go func() {
		writeToFile(recordChan)
	}()

	numGoroutines := runtime.NumCPU() * factor
	chunkSize := numRecords / numGoroutines

	metricNumberOfCPU.Set(float64(runtime.NumCPU()))
	metricNumberOfGoroutines.Set(float64(numGoroutines))
	metricChunkSize.Set(float64(chunkSize))

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(startIndex int) {
			defer wg.Done()

			for j := startIndex; j < startIndex+chunkSize; j++ {
				employment := Employment{
					Company:   fmt.Sprintf("Company %d", j+1),
					Position:  "Software Engineer",
					StartDate: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					EndDate:   time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
					Responsibilities: []string{
						"Developed and maintained web applications",
						"Collaborated with cross-functional teams",
						"Participated in code reviews",
					},
				}
				recordChan <- employment
			}
		}(i * chunkSize)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	go PrintMemUsage(ctx)
	wg.Wait()
	cancelFn()
}
func PrintMemUsage(ctx context.Context) {
	gcTicker := time.NewTicker(10 * time.Second)
	defer gcTicker.Stop()

	printTicker := time.NewTicker(2 * time.Second)
	defer printTicker.Stop()

	for {
		select {
		case <-gcTicker.C:
			log.Println("Force GC")
			runtime.GC()
			metricNumberOfGC.Inc()
		case <-printTicker.C:
			var memStats runtime.MemStats

			runtime.ReadMemStats(&memStats)
			log.Printf("Actual number of goroutines: %d\n", runtime.NumGoroutine())
			metricActualNumberOfGoroutines.Set(float64(runtime.NumGoroutine()))
			log.Printf("Alloc = %v MiB", memStats.Alloc/1024/1024)
			metricAllocatedMemory.Observe(float64(memStats.Alloc))
			log.Printf("TotalAlloc = %v MiB", memStats.TotalAlloc/1024/1024)
			log.Printf("Sys = %v MiB", memStats.Sys/1024/1024)
			log.Printf("NumGC = %v\n", memStats.NumGC)
		case <-ctx.Done():
			log.Println("Context cancelled, stop printing memory usage")
			return
		}
	}
}

func writeToFile(recordChan chan Employment) {

	// jsonFile, err := os.Create("employment_records.json")
	// if err != nil {
	// 	log.Fatal("Error creating JSON file:", err)

	// 	return
	// }
	jsonFile := new(noOpsWriterCloser)
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "    ")

	for record := range recordChan {
		err := encoder.Encode(record)
		if err != nil {
			log.Fatal("Error encoding JSON:", err)

			return
		}
	}

	log.Println("All employment records have been generated and saved to employment_records.json")
}

type noOpsWriterCloser struct {
	io.WriteCloser
}

func (noOpsWriterCloser) Close() error { return nil }
func (noOpsWriterCloser) Write(p []byte) (int, error) {
	return len(p), nil
}
