
FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o main .

# Deploy the application
CMD ["./main"]
