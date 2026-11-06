FROM golang:1.27-alpine AS build
WORKDIR /src
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D app
USER app
WORKDIR /app
COPY --from=build /out/ /app/
EXPOSE 8080
CMD ["/app/api"]