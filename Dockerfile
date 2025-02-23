FROM docker.io/golang:1.23-alpine AS build

WORKDIR /app
COPY . ./
RUN go mod tidy \
    && go build -ldflags="-s -w" -v -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main ./
RUN chmod +x ./main
ENTRYPOINT ["./main"]