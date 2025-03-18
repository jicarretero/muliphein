# Patch Integration Tests of Canis Major

In order to pass the Integration tests of Canis Major, after downloading the Canis Major from GIT

```
git clone https://github.com/FIWARE/CanisMajor.git
```

Here by I provide a patch to Canis Major testing (the default version is very restrictive with IPs / services), so 

```
cd CanisMajor
patch -p1 < ${MULIPHEIN_CODE_HOME}/muliphein/CanisMajorPatch/p1.patch
```

Now, for testing, you'll need to set up a set of variables:

```
# The following variable is not needed, just a commodity variable, to query Orion-LD (<host>:<port>)
export ORION_ADDRESS=orion_host:1026

# The next variable should point to the muliphein/cm-forwarder-proxy (<host>:<port>) or any proxy used.
# If there is no proxy, then CANIS_MAJOR_ADDRESS is OK
export NGSI_ADDRESS=muliphein_host:8080

# This variable points to the CANIS-MAJOR host (<host>:<port>)
export CANIS_MAJOR_ADDRESS=canis_major_host_8080

# Vault installation (<host>:<port>)
export VAULT_ADDRESS=vault_host:8200

# Ethereum URL (ganache in our case). Note this is an URL, not a host:port 
export ETHEREUM_URL="http://ganache-cli-uop-fiware-canis-major.apps.p2code-testbed.rh-horizon.eu";
```
