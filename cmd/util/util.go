package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

type Employment struct {
	Company          string
	Position         string
	StartDate        time.Time
	EndDate          time.Time
	Responsibilities []string
}

func main() {
	var wg sync.WaitGroup

	recordChan := make(chan Employment, 1000)
	defer close(recordChan)

	go func() {
		writeToFile(recordChan)
	}()

	numRecords := 10000000

	numGoroutines := runtime.NumCPU() * 4
	log.Printf("Number of CPU: %d\n", runtime.NumCPU())
	log.Printf("Number of goroutines: %d\n", numGoroutines)

	chunkSize := numRecords / numGoroutines

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
		case <-printTicker.C:
			var memStats runtime.MemStats

			runtime.ReadMemStats(&memStats)
			log.Printf("Actual number of goroutines: %d\n", runtime.NumGoroutine())
			log.Printf("Alloc = %v MiB", memStats.Alloc/1024/1024)
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

	jsonFile, err := os.Create("employment_records.json")
	if err != nil {
		log.Fatal("Error creating JSON file:", err)

		return
	}
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

	fmt.Println("All employment records have been generated and saved to employment_records.json")
}
