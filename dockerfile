FROM golang:alpine as builder-service

# ENV для Golang
ENV GO111MODULE=on

WORKDIR /go/src/app

COPY . .

RUN go mod tidy

RUN go build -o ./run ./app

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Yekaterinburg
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app/

COPY --from=builder-service /go/src/app/run .
COPY --from=builder-service /go/src/app/internal/templates/ .
COPY --from=builder-service /go/src/app/internal/images/ .

EXPOSE 8080

ENTRYPOINT ["./run"]