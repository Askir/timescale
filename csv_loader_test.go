package main

import (
	"os"
	"testing"
	"time"
)

func TestReadQueryParams(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test_query_params_*.csv")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		testData := `hostname,start_time,end_time
host_1,2023-01-01 00:00:00,2023-01-01 01:00:00
host_2,2023-01-01 01:00:00,2023-01-01 02:00:00
host_3,2023-01-01 02:00:00,2023-01-01 03:00:00`
		if _, err := tempFile.WriteString(testData); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		params, err := LoadQueryParams(tempFile.Name())

		if err != nil {
			t.Fatalf("readQueryParams returned an error: %v", err)
		}

		if len(params) != 3 {
			t.Errorf("Expected 3 params, got %d", len(params))
		}

		expected := QueryParams{
			Hostname:  "host_1",
			StartTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC),
		}
		if params[0] != expected {
			t.Errorf("First param mismatch. Expected %+v, got %+v", expected, params[0])
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		_, err := LoadQueryParams("non_existent_file.csv")
		if err == nil {
			t.Error("Expected an error for non-existent file, got nil")
		}
	})

	t.Run("Permission denied", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test_no_permissions_*.csv")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		if err := os.Chmod(tempFile.Name(), 0000); err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}

		_, err = LoadQueryParams(tempFile.Name())
		if err == nil {
			t.Error("Expected an error for file with no read permissions, got nil")
		}
	})

	t.Run("Malformed CSV", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test_malformed_*.csv")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		malformedTestData := `hostname,start_time,end_time
host_1,2023-01-01 00:00:00,2023-01-01 01:00:00
host_2,invalid_date,2023-01-01 02:00:00
host_3,2023-01-01 02:00:00,invalid_date`
		if _, err := tempFile.WriteString(malformedTestData); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		_, err = LoadQueryParams(tempFile.Name())
		if err == nil {
			t.Error("Expected an error for malformed CSV data, got nil")
		}
	})
}
