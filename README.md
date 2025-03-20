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

Sometimes, things doesn't go as expected, in order to check the requests it receives, it can write `curl`commands equivalent to requets it receives (it shows localhost:1026 instead of current IPs).

These curl commands will be writen to files named `/tmp/here-x.req`, where x is a sequential number starting from 1. It also shows these commands in logs.

```
export  DUMP_AS_CURL=yes
```

## Decisions

Of course, if I forward the requests to 2 different servers, the responses can be different. So, what one will be returned?. Here some decisions have been adopted, and these decisions are opinionated, maybe this change in the future.

- On POSTS / PATCH / PUT / DELETE - Resposes are returned from Canis majore (or whatever CANIS_MAJOR_URL is). In case of 40X or 50x responses, it will return the ngsi-ld broker response (or whatever it is NGSILD_BROKER_URL).

- On GET / HEAD - It won't forward anything to Canis Major (CANIS_MAJOR_URL) and it will return whatever the NGSI-LD Broker response (NGSILD_BROKER_URL).

This piece of software also solves a glitch between the Canis Major and OrionLD. At this moment, there are some know issues with Orion LD POSTS to `.../attrs` and Canis Major doesn't support PATCH to `.../attrs`, so, the following
rule applies:

- PATCH to `.../attrs`will be used as a PATCH. Nothing changes.
- POST to `.../attrs`will be used as a POST to CANIS_MAJOR and as PATCH to NGSILD_BROKER.

I also changed the integration tests for Canis Major and here I provide the patch. Hopefully in the future there will be a proper Pull Request to Canis Major. The patch is described in [CanisMajorPatch](CanisMajorPatch/README.md).

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
