

FROM node:15.5.1-alpine3.12 AS nodeBuilder
WORKDIR /app
COPY ./web ./
ENV PATH /app/node_modules/.bin:$PATH
RUN npm install
RUN npm run build

FROM golang:1.15 as goBuilder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -ldflags "-s -w" -v -o server

FROM alpine:3
RUN apk add --no-cache ca-certificates
COPY --from=goBuilder /app/server /server
COPY --from=goBuilder /app/variables.json /.
COPY --from=nodeBuilder /app/build /web/build/
CMD ["/server"]