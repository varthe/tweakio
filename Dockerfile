FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o tweakio ./cmd/main.go

FROM gcr.io/distroless/static-debian11 AS runner

WORKDIR /app

COPY --from=builder /app/tweakio /app/tweakio

EXPOSE 3185

CMD ["/app/tweakio"]