
# BUILD THE GO APPLICATION IMAGE 

FROM golang:1.27.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /ecommerce-server ./cmd/server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates \
        && addgroup -S appgroup \
        && adduser -S -D -H -G appgroup appuser \
        && mkdir -p /app/logs \
        && chown -R appuser:appgroup /app

COPY --from=builder --chown=appuser:appgroup /ecommerce-server /app/ecommerce-server

USER appuser

EXPOSE 8080

CMD ["./ecommerce-server"]
