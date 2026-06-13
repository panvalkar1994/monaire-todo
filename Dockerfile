FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /monaire-todo .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /monaire-todo .
COPY migrations ./migrations
COPY config/ ./config/
RUN cp config/config.docker.yaml config/config.yaml
EXPOSE 8080
ENTRYPOINT ["./monaire-todo"]
