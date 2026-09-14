# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o media-sequencer .

# ---- Final stage ----
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/media-sequencer .

EXPOSE 8080

CMD ["./media-sequencer"]
