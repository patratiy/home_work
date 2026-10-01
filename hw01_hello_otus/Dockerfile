FROM golang:1.26

WORKDIR /usr/src/app

# Pre-copy go.mod for dependency caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]