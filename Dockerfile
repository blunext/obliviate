FROM node:23-alpine AS nodeBuilder
WORKDIR /app
COPY ./web ./
ENV PATH /app/node_modules/.bin:$PATH
RUN npm install
RUN npm run build

FROM golang:1.25 as goBuilder
RUN useradd -u 10001 -d /app scratchuser
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
COPY --from=nodeBuilder /app/build /app/web/build/
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -ldflags "-s -w" -v -o server

FROM scratch
COPY --from=goBuilder /app/server /server
COPY --from=goBuilder /etc/passwd /etc/passwd
USER scratchuser
CMD ["/server"]