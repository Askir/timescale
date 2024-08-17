# CPU Usage Query Benchmark

This project provides a Go application that queries CPU usage data from a TimescaleDB database and generates benchmark statistics based on query time.

## Prerequisites

- Docker
- Docker Compose
- Go 1.23 or later (for running tests locally)

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/Askir/timescale.git
   cd timescale
   ```

2. Start the Docker containers:
   ```
   docker-compose up -d
   ```
   This will start the TimescaleDB container and build and run the application container.

## Running Tests

To run the tests locally:

1. Ensure you have Go 1.23 or later installed on your machine.

3. Run the following command:
   ```
   go test
   ```

This will run all tests in the project.

## Project Structure

- `main.go`: The main application logic. Containing the distribution to workers and actual query execution as well as the statistics calculation.
- `csv_loader.go`: Functions for loading query parameters from the CSV file.

## Notes

- The application uses environment variables for configuration. Make sure to set `WORKER_COUNT` and `DATABASE_URL` appropriately (they are set in the `docker-compose.yaml`).
- The TimescaleDB data is persisted in the `./data` directory.
