package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/montanaflynn/stats"
	dataframe "github.com/rocketlaunchr/dataframe-go"
	"github.com/spf13/cobra"
)

// Compile regex used to extract the host number from its string
var hostRegex = regexp.MustCompile(`host_(\d{6})`)

// query represents each row from the input CSV file
type query struct {
	host      string
	startTime time.Time
	endTime   time.Time
}

// row represents each row from the result set returned by executing
// the query. It's used by an auxiliary database client library to save the
// values from the result set into structs
type row struct {
	Time time.Time
	Max  float64
	Min  float64
}

// main validates the command line arguments and calls run()
func main() {
	var file string
	var workers int

	cmd := &cobra.Command{
		Use:   "querybench",
		Short: "querybench is a SQL query benchmarking tool",
		Run: func(cmd *cobra.Command, args []string) {
			run(file, workers)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to query file")
	cmd.Flags().IntVarP(&workers, "workers", "w", 8, "Number of workers")

	cmd.MarkFlagRequired("file")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}

// run implements the query benchmark logic
func run(path string, workerCount int) {
	var mutex sync.Mutex // Use a mutex to coordinate writing to the results series
	results := dataframe.NewSeriesFloat64("Time", nil)

	// Each worker runs  its own goroutine, consumes its own channel and opens
	// its own database connection
	workers := make([]chan *query, workerCount)
	for i := 0; i < workerCount; i++ {
		workers[i] = make(chan *query)
	}

	var waitWorkers sync.WaitGroup

	// For each worker, start a goroutine, connect to the database, and begin
	// consuming the corresponding query channel
	for _, workerChan := range workers {
		waitWorkers.Add(1)

		go func(channel chan *query) {
			conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
			if err != nil {
				panic(err)
			}

			// Consume the query channel until the channel is closed
			for {
				query, ok := <-channel
				if !ok {
					break
				}

				execTime := executeQuery(conn, query)

				mutex.Lock()
				results.Append(execTime.Milliseconds())
				mutex.Unlock()
			}

			conn.Close(context.Background())
			waitWorkers.Done()
		}(workerChan)
	}

	// Open the input file, parse each row into a query struct, and send it to
	// the query channel associated with its hostname
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	reader := csv.NewReader(file)

	reader.Read() // Skip header row

	var waitRead sync.WaitGroup

	// For each row, parse it to a query struct, find the associated worker, and
	// send the query to its channel
	for {
		row, err := reader.Read()
		if row == nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		query := parseRow(row)
		worker := mapWorker(query, workerCount)

		waitRead.Add(1)
		go func() {
			workers[worker] <- query
			waitRead.Done()
		}()
	}

	waitRead.Wait()

	// Stop workers
	for _, workerChan := range workers {
		close(workerChan)
	}

	waitWorkers.Wait()

	// Compute and print results
	median, _ := stats.Median(results.Values)
	mean, _ := stats.Mean(results.Values)
	min, _ := stats.Min(results.Values)
	max, _ := stats.Max(results.Values)

	fmt.Printf("Total: %d\n", results.NRows())
	fmt.Printf("Median: %.2f\n", median)
	fmt.Printf("Mean: %.2f\n", mean)
	fmt.Printf("Min: %.2f\n", min)
	fmt.Printf("Max: %.2f\n", max)
}

// panic prints an error message to stderr and exits.
func panic(err error) {
	fmt.Printf("Error: %s\n", err)
	os.Exit(1)
}

// mapWorker pins a host to a specific worker to accomodate the constraint of
// same host queries being executed by the same worker
func mapWorker(query *query, poolSize int) int {
	matches := hostRegex.FindStringSubmatch(query.host)
	worker, err := strconv.Atoi(matches[1])
	if err != nil {
		panic(err)
	}

	return worker % poolSize
}

// parseRow converts a CSV field array into a query struct
func parseRow(row []string) *query {
	start, err := time.Parse("2006-01-02 15:04:05", row[1])
	if err != nil {
		panic(err)
	}

	end, err := time.Parse("2006-01-02 15:04:05", row[2])
	if err != nil {
		panic(err)
	}

	return &query{
		host:      row[0],
		startTime: start,
		endTime:   end,
	}
}

// executeQuery runs the benchmark query and measures the returns its execution
// time
func executeQuery(conn *pgx.Conn, query *query) time.Duration {
	var rows []*row

	start := time.Now()
	err := pgxscan.Select(context.Background(), conn, &rows, `
		SELECT time_bucket('1 minute', ts) as time, max(usage), min(usage)
		FROM cpu_usage
		WHERE host = $1 AND ts >= $2 AND ts <= $3
		GROUP BY time
	`, query.host, query.startTime, query.endTime)
	if err != nil {
		panic(err)
	}
	execTime := time.Since(start)

	return execTime
}
