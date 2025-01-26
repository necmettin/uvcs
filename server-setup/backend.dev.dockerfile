FROM golang:1.21-alpine

WORKDIR /app

# Install Air for hot reloading and other necessary tools
RUN go install github.com/cosmtrek/air@v1.44.0 && \
    apk add --no-cache curl

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Create the .air.toml configuration
RUN air init

# Update .air.toml for our needs
RUN sed -i 's/cmd = "go run \.\/main\.go"/cmd = "go run main.go"/' .air.toml && \
    sed -i 's/exclude_dir = \["assets", "tmp", "vendor", "testdata"\]/exclude_dir = \["assets", "tmp", "vendor", "testdata", "data"\]/' .air.toml

# Expose port
EXPOSE 8080

# Start Air for hot reloading
CMD ["air", "-c", ".air.toml"] 