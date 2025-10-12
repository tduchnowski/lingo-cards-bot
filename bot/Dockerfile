FROM golang:latest AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o lingo-cards-bot ./cmd

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=build /build/lingo-cards-bot .
COPY --from=build /build/cmd/config.yaml ./cmd/
CMD ["/app/lingo-cards-bot"]
