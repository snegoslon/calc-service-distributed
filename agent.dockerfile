FROM golang:1.23.1-alpine
WORKDIR /app
COPY . .
ENV COMPUTING_POWER=4 \
    ORCHESTRATOR_URL=http://orchestrator:8080
RUN go build -o agent ./cmd/agent
CMD ["./agent"]
