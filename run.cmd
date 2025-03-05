set TIME_ADDITION_MS=200
set TIME_SUBTRACTION_MS=200
set TIME_MULTIPLICATIONS_MS=300
set TIME_DIVISIONS_MS=400

start cmd /k "go run ./cmd/orchestrator/main.go"

set COMPUTING_POWER=4
set ORCHESTRATOR_URL=http://localhost:8080

start cmd /k "go run ./cmd/agent/main.go"
