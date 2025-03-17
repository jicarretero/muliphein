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

It listens on port 8080

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
    -p 4343:8080 muliphein:0.5
```

