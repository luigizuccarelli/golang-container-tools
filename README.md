## Overview

This is a simple POC that converts v1 image (from a registry) to OCI format
copying all blobs and manifests to a local directory

There is also a push feature that takes all data (OCI format) from the local directory
and pushes to a compliant OCI registry

## Use Case

A typical use case could be:
- Work with catalog i.e registry.redhat.io/redhat/redhat-operator-index:v4.11
- Copy to local directory in OCI format (Catalog and related images)
- Make changes locally
- Push to OCI compliant registry

## POC 

I used a simple approach - Occam's razor

- A scientific and philosophical rule that entities should not be multiplied unnecessarily (KISS)
- Worked with a v1 image for the POC
- Add the catalog and relevant images in a catalog later


## Usage

Execute the following to copy from a registry

```bash
./build/oci -a copy -i quay.io/<user>/<image-name> -v v0.0.1 -p test-oci -t true -b false

# parameters
  -a is action i.e copy or push
  -i image url
  -v version
  -p local path
  -t tls-verify (true or false)
  -b basic auth (true or false)
```

Execute the following to push to a registry

```bash
./build/oci -a push -i localhost:5000/<image-name> -v v0.0.1 -p test-oci -t false -b true

# parameters (see above)

```

## Building

The project uses a Makefile

Binary will be in the folder *build/*

```bash

make build
```
