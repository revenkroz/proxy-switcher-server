FROM golang:1.21-alpine AS base

WORKDIR /app

# builder
FROM base AS builder

# build-args
ARG VERSION

RUN apk --no-cache add bash git openssh

# modules: utilize build cache
COPY go.mod ./

RUN go mod download
COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o proxy-switcher -ldflags "-w -s" .

# runner
FROM scratch AS prod

WORKDIR /app

COPY --from=builder /app/proxy-switcher /app/proxy-switcher
EXPOSE 8888

CMD [ "/app/proxy-switcher" ]
