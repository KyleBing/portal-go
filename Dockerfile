# syntax=docker/dockerfile:1

FROM node:20-alpine AS frontend
WORKDIR /src
COPY web/manager/package.json web/manager/yarn.lock* web/manager/package-lock.json* ./
RUN yarn install --frozen-lockfile || yarn install
COPY web/manager/ ./
RUN yarn build

FROM golang:1.22-alpine AS backend
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/dist ./web/manager/dist
RUN CGO_ENABLED=0 go build -o /out/portal ./cmd/portal \
 && CGO_ENABLED=0 go build -o /out/ws ./cmd/ws \
 && CGO_ENABLED=0 go build -o /out/cron ./cmd/cron

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /out/portal /out/ws /out/cron /app/
COPY config /app/config
COPY migrations /app/migrations
COPY --from=backend /src/web/manager/dist /app/web/manager/dist
RUN mkdir -p /app/upload /app/temp
ENV PORT=3000
EXPOSE 3000 9999
CMD ["/app/portal"]
