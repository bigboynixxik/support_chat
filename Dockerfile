FROM golang:1.25.0-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/support_chat ./cmd/app/main.go

FROM alpine:3.21

WORKDIR /app
RUN apk add --no-cache ca-certificates git

COPY --from=builder /out/support_chat /app/support_chat


EXPOSE 8082

CMD ["/app/support_chat"]