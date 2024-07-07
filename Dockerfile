FROM golang:1.22.1 as build

ARG PACKAGE

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY . .
RUN go mod download

RUN cd "$PACKAGE" && CGO_ENABLED=0 GOOS=linux go build -o main

FROM golang:1.22.1

ARG PACKAGE

WORKDIR /app

COPY --from=build "/app/$PACKAGE/main" main

EXPOSE 8080

CMD ["/app/main"]