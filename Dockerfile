FROM golang:1.22.1 as build

ARG PACKAGE

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY . .
RUN go mod download

WORKDIR "/app/$PACKAGE"

RUN CGO_ENABLED=0 GOOS=linux go build -o main

EXPOSE 8080

CMD ["./main"]