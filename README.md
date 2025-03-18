# muliphein

γ Canis Majoris: This is Somehow in the middle of Canis Major and NGSI-LD brokers.

This is a HTTP web server which accepts incoming requests and forks them to two servers, ideally a FIWARE NGISLD-Broker and a  FIWARE Canis Major. In fact, it is tought to run like that.

It is mandatory 2 environment variables:

- CANIS_MAJOR_URL
- NGSILD_BROKER_URL

They must be defined or the program will die. example:

```
export CANIS_MAJOR_URL=http://localhost:4000
export NGSILD_BROKER_URL=htpt://localhost:1026
```

It listens on port **8080**, it can't be configured.

## Help debugging

Sometimes, things doesn't go as expected, in order to check the requests it receives, it can wirte `curl`commands equivalent to requets it receives (it shows localhost:1026 instead of current IPs).

These curl commands will be writen to files named `/tmp/here-x.req`, where x is a sequential number starting from 1. It also shows these commands in logs.

```
export  DUMP_AS_CURL=yes
```

## Decisions

Of course, if I forward the requests to 2 different servers, the responses can be different. So, what one will be returned?. Here some decisions have been adopted, and these decisions are opinionated, maybe this change in the future.

- On POSTS / PATCH / PUT - Resposes are returned from Canis majore (or whatever CANIS_MAJOR_URL is. In case of 40X or 50x responses, it will return the NGSILD_BROKER_URL.

- On GET / HEAD / DELETE - It won't forward anything to CANIS_MAJOR and it will return whatever the NGSILD_BROKER returns.

## Dockerfile

Podemos crear el docker de **muliphein** simplemente ejecutando

```
docker build -t muliphein:${version} .
```

Posteriormente, podemos ejecutarlo (sea esto un ejemplo) tal que así:

```
docker run --rm \
    -e NGSILD_BROKER_URL=http://192.168.3.253:1026 \
    -e CANIS_MAJOR_URL=http://192.168.3.253:1036 \
    -p 4343:8080 muliphein:${version}
```
