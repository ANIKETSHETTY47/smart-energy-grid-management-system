FROM golang:1.22 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
