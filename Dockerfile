

FROM node:15.5.1-alpine3.12 AS nodeBuilder
WORKDIR /app
COPY ./web ./
ENV PATH /app/node_modules/.bin:$PATH
RUN npm install
RUN npm run build

FROM golang:latest as goBuilder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
COPY --from=nodeBuilder /app/build /app/web/build/
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -ldflags "-s -w" -v -o server

FROM alpine:3
RUN apk add --no-cache ca-certificates
COPY --from=goBuilder /app/server /server
CMD ["/server"]