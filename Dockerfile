FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o crowdsay ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/crowdsay .
COPY .env .
EXPOSE 8080
CMD ["./crowdsay"]
