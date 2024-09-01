# Use the official Go image as a parent image
FROM golang:1.22-alpine

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

ENV CGO_ENABLED=1

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8000

# Command to run the executable
CMD ["./main"]
