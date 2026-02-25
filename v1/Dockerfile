FROM golang:alpine3.23

WORKDIR /src

COPY . .

RUN go mod tidy

EXPOSE 8080
CMD [ "go", "run", "." ]