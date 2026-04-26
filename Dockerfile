FROM node:22-alpine AS admin-builder

WORKDIR /src/admin

RUN corepack enable

COPY admin/package.json admin/pnpm-lock.yaml admin/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY admin/ ./
RUN pnpm build

FROM --platform=$BUILDPLATFORM golang:1.26.1-alpine AS go-builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/go-csust-planet ./main.go

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

ENV APP_MODE=production
ENV TZ=Asia/Shanghai
ENV PORT=7241

WORKDIR /app

RUN mkdir -p /app/secrets

COPY --from=go-builder /out/go-csust-planet /app/go-csust-planet
COPY --from=admin-builder /src/admin/build /app/admin/build

EXPOSE 7241

CMD ["/app/go-csust-planet"]
