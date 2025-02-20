FROM docker.io/golang:1.23

WORKDIR /app
COPY . ./
RUN go mod tidy \
    && go build -ldflags="-s -w" -v -o main . \
    && chmod +x main

ENTRYPOINT ["./main"]