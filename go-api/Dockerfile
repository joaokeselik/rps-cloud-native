FROM golang:1.22 AS build

WORKDIR /src

COPY go-api/go.mod go-api/go.sum* ./
RUN go mod download

COPY go-api ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /players-api ./cmd/server

FROM scratch

WORKDIR /
COPY --from=build /players-api /players-api
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/players-api"]
