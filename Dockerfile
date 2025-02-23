FROM docker.io/golang:1.23 AS build

WORKDIR /app
COPY . ./
RUN go mod tidy \
    && go build -ldflags="-s -w" -v -o main . \
    && chmod +x main

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .
ENTRYPOINT ["./main"]