# Build stage
FROM --platform=linux/amd64 golang:latest AS builder
WORKDIR /app

# Declare the build argument
ARG DATABASE_NAME

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

# Final stage
FROM --platform=linux/amd64 alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Set environment variable from build argument
ARG DATABASE_NAME
ENV DATABASE_NAME=${DATABASE_NAME}


COPY --from=builder /app/app .
RUN echo "DATABASE_NAME=${DATABASE_NAME}" > .env
EXPOSE 8080
CMD ["./app"]