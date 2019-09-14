FROM golang:1.13-alpine
WORKDIR /app
COPY . . 
RUN go mod download
RUN go build -o app .
CMD ["./app", "--tcpport=8080", "--tcpaddr=0.0.0.0"]