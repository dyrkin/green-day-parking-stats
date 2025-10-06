FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /app/app .
EXPOSE 2112
ENTRYPOINT ["/app"]