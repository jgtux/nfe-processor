FROM golang:1.22-alpine AS builder

RUN apk add --no-cache libxml2-dev gcc musl-dev

WORKDIR /app

COPY . .
RUN CGO_ENABLED=1 go build -o nfe-server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache libxml2

WORKDIR /app
COPY --from=builder /app/nfe-server .
COPY internal/parser/schemas/ ./schemas/

EXPOSE 8080
CMD ["./nfe-server"]
