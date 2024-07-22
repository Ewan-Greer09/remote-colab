# Use the official Golang image as the base image
FROM golang:1.22

RUN go install github.com/air-verse/air@latest
RUN go install github.com/a-h/templ/cmd/templ@v0.2.680

# Install npm and tailwindcss
RUN apt-get update && apt-get install -y npm \
    && apt-get clean
RUN npm install -g tailwindcss

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

ENV PORT=3000
EXPOSE 3000

RUN templ generate
RUN npx tailwindcss -i ./internal/public/css/index.css -o ./internal/public/css/output.css

ENTRYPOINT ["go", "run", "./..."]