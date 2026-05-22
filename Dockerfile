# Cloud Run job container image for this Fido fetch project.
#
# Build stage.
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/fetch ./cmd/fetch

# Runtime stage.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/fetch /usr/local/bin/fetch
ENTRYPOINT ["/usr/local/bin/fetch"]
