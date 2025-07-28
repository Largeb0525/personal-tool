FROM golang:1.22
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY config/local.toml config/config.toml
COPY . .
RUN go build -o main .
CMD ["/app/main"]