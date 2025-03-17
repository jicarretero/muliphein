FROM golang:alpine3.21 AS build

# update certificates to trust github
COPY ./main.go /tmp/main.go

WORKDIR /tmp
RUN go build -o muliphein ./main.go

FROM alpine:3.21
COPY --from=build /tmp/muliphein /usr/bin/muliphein
RUN  apk add wget curl bind-tools net-tools

ENTRYPOINT [ "/usr/bin/muliphein" ]
