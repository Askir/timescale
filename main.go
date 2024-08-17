package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

type Connection interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Close(context.Context) error
}

var connect = func(ctx context.Context, connString string) (Connection, error) {
	return pgx.Connect(ctx, connString)
}

var execQuery = func(conn Connection, ctx context.Context, id int, query string, hostname string, startTime time.Time, endTime time.Time) {
	rows, err := conn.Query(ctx, query, hostname, startTime, endTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Worker %d: Query error for host %s: %v\n", id, hostname, err)
	} else {
		rows.Close()
	}
}

type QueryParams struct {
	Hostname  string
	StartTime time.Time
	EndTime   time.Time
}

type QueryResult struct {
	Hostname string
	Duration time.Duration
}

const query = `
	SELECT 
		date_trunc('minute', ts) AS minute,
		MAX(usage) AS max_usage,
		MIN(usage) AS min_usage
	FROM cpu_usage
	WHERE host = $1 AND ts >= $2 AND ts < $3
	GROUP BY minute
	ORDER BY minute
`

func main() {
	params, err := LoadQueryParams("query_params.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading query params: %v\n", err)
		os.Exit(1)
	}

	results, overallDuration := executeQueriesInParallel(params)
	durations := processStatistics(results)

	printStatistics(durations, overallDuration)
}

func executeQueriesInParallel(params []QueryParams) (chan QueryResult, time.Duration) {
	results := make(chan QueryResult, len(params))
	var wg sync.WaitGroup

	numWorkers, err := strconv.Atoi(os.Getenv("WORKER_COUNT")) // Adjust as needed
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading WORKER_COUNT: %v\n", err)
		os.Exit(1)
	}
	workerChannels := make([]chan QueryParams, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerChannels[i] = make(chan QueryParams, 100)
		go worker(i, &wg, results, workerChannels[i])
	}

	overallStart := time.Now()
	for i, p := range params {
		wg.Add(1)
		workerIndex := getWorkerIndex(p.Hostname, numWorkers)
		workerChannels[workerIndex] <- p
		if i > 0 && i%100 == 0 {
		}
	}

	for _, ch := range workerChannels {
		close(ch)
	}

	wg.Wait()
	close(results)
	overallDuration := time.Since(overallStart)
	return results, overallDuration
}

func processStatistics(results <-chan QueryResult) []time.Duration {
	var durations []time.Duration
	var totalDuration time.Duration
	resultCount := 0
	for r := range results {
		durations = append(durations, r.Duration)
		totalDuration += r.Duration
		resultCount++
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	return durations
}

func printStatistics(durations []time.Duration, overallDuration time.Duration) {
	var totalDuration time.Duration
	for _, d := range durations {
		totalDuration += d
	}

	numQueries := len(durations)
	minDuration := durations[0]
	maxDuration := durations[numQueries-1]
	medianDuration := durations[numQueries/2]
	avgDuration := totalDuration / time.Duration(numQueries)

	// Print benchmark stats
	fmt.Println("Result of benchmark:")
	fmt.Printf("Overall execution time: %v\n", overallDuration)
	fmt.Printf("Number of queries: %d\n", numQueries)
	fmt.Printf("Total query processing time: %v\n", totalDuration)
	fmt.Printf("Minimum query time: %v\n", minDuration)
	fmt.Printf("Median query time: %v\n", medianDuration)
	fmt.Printf("Average query time: %v\n", avgDuration)
	fmt.Printf("Maximum query time: %v\n", maxDuration)

	fmt.Println("Application finished.")
}

func getWorkerIndex(hostname string, numWorkers int) int {
	h := fnv.New32a()
	h.Write([]byte(hostname))
	return int(h.Sum32()) % numWorkers
}

func worker(id int, wg *sync.WaitGroup, results chan<- QueryResult, params <-chan QueryParams) {
	conn, err := connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Worker %d: Unable to connect to database: %v\n", id, err)
		return
	}
	defer conn.Close(context.Background())

	for p := range params {
		start := time.Now()
		execQuery(conn, context.Background(), id, query, p.Hostname, p.StartTime, p.EndTime)
		duration := time.Since(start)
		results <- QueryResult{Hostname: p.Hostname, Duration: duration}
		wg.Done()
	}
}
