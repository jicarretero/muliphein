FROM golang:1.24-alpine3.21 AS build

# update certificates to trust github
RUN mkdir /tmp/muliphein
COPY . /tmp/muliphein

WORKDIR /tmp/muliphein
RUN go build -o muliphein .

FROM alpine:3.21
COPY --from=build /tmp/muliphein/muliphein /usr/bin/muliphein
RUN  apk add wget curl bind-tools net-tools

ENTRYPOINT [ "/usr/bin/muliphein" ]
