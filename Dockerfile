#build stage
FROM golang:alpine AS builder
WORKDIR /app/godockerexample
COPY . .
RUN go mod download
RUN go build -o /godocker
EXPOSE 8081
CMD [ "/godocker" ]