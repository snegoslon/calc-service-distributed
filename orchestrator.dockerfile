FROM golang:1.23.1-alpine
WORKDIR /app
COPY . .
ENV TIME_ADDITION_MS=200 \
    TIME_SUBTRACTION_MS=200 \
    TIME_MULTIPLICATIONS_MS=300 \
    TIME_DIVISIONS_MS=400
RUN go build -o orchestrator ./cmd/orchestrator
CMD ["./orchestrator"]
