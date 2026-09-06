FROM node:24-alpine AS node-builder
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.27 AS go-builder
RUN useradd -u 10001 -d /app scratchuser
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
COPY --from=node-builder /app/build /app/web/build/
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -trimpath -ldflags "-s -w" -o server

FROM scratch
COPY --from=go-builder /app/server /server
COPY --from=go-builder /etc/passwd /etc/passwd
USER scratchuser
CMD ["/server"]
