FROM golang:1.24.2-alpine

WORKDIR /app
COPY . .

# Run the application
CMD ["go", "run", "./cmd/main.go"]