FROM golang:1.13-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./main.go ./
RUN go build -o app .
CMD ["./app", "--tcpport=8080", "--tcpaddr=0.0.0.0"]