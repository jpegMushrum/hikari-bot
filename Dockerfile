FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
# load deps
RUN apk update --no-cache && apk add --no-cache tzdata
WORKDIR /build
ADD go.mod .
ADD go.sum .
RUN go mod download

# build
COPY . .
RUN go build -o /app/bot 

# copy binary
FROM alpine

WORKDIR /app

COPY --from=builder /app/bot /app/bot
CMD ["/app/bot"]
