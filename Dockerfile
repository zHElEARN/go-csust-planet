FROM --platform=$BUILDPLATFORM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/go-csust-planet ./main.go

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai
ENV PORT=7241

WORKDIR /app

RUN mkdir -p /app/secrets

COPY --from=builder /out/go-csust-planet /app/go-csust-planet

EXPOSE 7241

CMD ["/app/go-csust-planet"]
