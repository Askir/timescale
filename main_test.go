package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Close(ctx context.Context) error {
	return nil
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	called := m.Called(ctx, sql, args)
	return called.Get(0).(pgx.Rows), called.Error(1)
}

type ExecQueryCall struct {
	Conn      Connection
	Ctx       context.Context
	ID        int
	Query     string
	Hostname  string
	StartTime time.Time
	EndTime   time.Time
}

type MockExecQuery struct {
	mu    sync.Mutex
	Calls []ExecQueryCall
}

func (m *MockExecQuery) ExecQuery(conn Connection, ctx context.Context, id int, query string, hostname string, startTime time.Time, endTime time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, ExecQueryCall{conn, ctx, id, query, hostname, startTime, endTime})

}

func TestExecuteQueriesInParallel(t *testing.T) {
	os.Setenv("WORKER_COUNT", "3")

	testParams := []QueryParams{
		{Hostname: "host1", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
		{Hostname: "host2", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
		{Hostname: "host3", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
		{Hostname: "host4", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
		{Hostname: "host5", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
	}

	mockDB := new(MockDB)
	originalConnect := connect
	connect = func(ctx context.Context, connString string) (Connection, error) {
		return mockDB, nil
	}
	defer func() { connect = originalConnect }()

	mockExecQuery := &MockExecQuery{}
	originalExecQuery := execQuery
	execQuery = mockExecQuery.ExecQuery
	defer func() { execQuery = originalExecQuery }()

	results, overallDuration := executeQueriesInParallel(testParams)
	var receivedResults []QueryResult
	for result := range results {
		receivedResults = append(receivedResults, result)
	}

	assert.Equal(t, len(testParams), len(receivedResults), "Number of results should match number of input parameters")

	hostnamesProcessed := make(map[string]bool)
	for _, result := range receivedResults {
		hostnamesProcessed[result.Hostname] = true
	}
	for _, param := range testParams {
		assert.True(t, hostnamesProcessed[param.Hostname], "Hostname %s should have been processed", param.Hostname)
	}

	assert.Greater(t, overallDuration, time.Duration(0), "Overall duration should be greater than zero")

	assert.Equal(t, len(testParams), len(mockExecQuery.Calls), "execQuery should be called once for each input parameter")

	// Verify that execQuery was called with correct parameters
	for _, call := range mockExecQuery.Calls {
		found := false
		for _, param := range testParams {
			if param.Hostname == call.Hostname && param.StartTime.Equal(call.StartTime) && param.EndTime.Equal(call.EndTime) {
				found = true
				break
			}
		}
		assert.True(t, found, "ExecQuery call should match one of the input parameters")
		assert.Equal(t, query, call.Query, "Query should match expected SQL")
	}

	// Verify that worker distribution is correct
	workerCounts := make(map[int]int)
	for _, call := range mockExecQuery.Calls {
		workerCounts[call.ID]++
	}
	assert.Equal(t, 3, len(workerCounts), "Queries should be distributed across 3 workers")
}
