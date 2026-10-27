FROM golang:1.25-alpine AS development

RUN go install github.com/air-verse/air@v1.62.0
WORKDIR /app
COPY go.mod ./
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM scratch AS production
COPY --from=build /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
