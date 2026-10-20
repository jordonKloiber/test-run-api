# Build stage: compiles a static binary. Nothing from here ships in the final image.
FROM golang:1.27 AS builder
WORKDIR /app

# Copied separately from the rest of the source so `go mod download` is only
# re-run when dependencies actually change, not on every source edit.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/api

# Final stage: distroless has no shell, no package manager, nothing but what's
# needed to run a static binary.
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
