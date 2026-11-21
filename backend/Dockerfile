FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY . .
# Pure Go since the XSD validator moved to lestrrat-go/helium: no cgo, no
# libxml2, so the binary is static and the runtime image needs nothing extra.
RUN CGO_ENABLED=0 go build -o nfe-server ./cmd/server

FROM alpine:3.23

WORKDIR /app
COPY --from=builder /app/nfe-server .
COPY internal/parser/schemas/ ./schemas/

EXPOSE 8080
CMD ["./nfe-server"]
