# Use the official Golang image as the base image
FROM golang:1.22

RUN go install github.com/cosmtrek/air@latest
RUN go install github.com/a-h/templ/cmd/templ@latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Set the entrypoint command to run the application with Air
ENTRYPOINT ["air", "-c", ".air.toml"]