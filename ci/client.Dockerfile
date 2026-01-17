FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY gql ./gql
COPY data/sample-nodes.csv ./sample-nodes.csv

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/client ./cmd/client

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/client /usr/local/bin/client
COPY --from=builder /src/gql ./gql
COPY --from=builder /src/sample-nodes.csv ./sample-nodes.csv

CMD ["/usr/local/bin/client"]
