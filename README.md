# Minio downloader for init containers

For downloading data from minio to your container to populate your PV

## Parameters

* takes parameters as env vars
* for `DEST` if you are downloading all files from a dir, don't include that dir name in your `DEST` Value. For example. If you were downloading all the contents from `users/roles` do _not_ make your destination `/data/users/roles`. *Do* make it `/data/users`. The roles directory will be created as part of the download. 

## Run locally for testing
### Binary
To grab all files in a dir
```
export MINIO_URL="192.168.2.66:9000"
export SRC="vectorizer/roles"
export ACCESSKEY="LEB3JJ3OCBN4HTDIS5IZ"
export SECRETKEY="xs3fx83cMkV7Oh+6jlGGTt9kTmT5D6yoQLm9+L5X"
export DEST="/tmp/minio-tests/examples"
```

Or for a single file

```
export MINIO_URL="192.168.2.66:9000"
export SRC="vectorizer/roles"
export ACCESSKEY="LEB3JJ3OCBN4HTDIS5IZ"
export SECRETKEY="xs3fx83cMkV7Oh+6jlGGTt9kTmT5D6yoQLm9+L5X"
export DEST="/tmp/minio-tests/examples/roles"
```

### Docker

``` shell
docker run -ti -v /tmp:/tmp \
-e MINIO_URL="192.168.2.66:9000" \
-e ACCESSKEY="LEB3JJ3OCBN4HTDIS5IZ" \
-e SECRETKEY="xs3fx83cMkV7Oh+6jlGGTt9kTmT5D6yoQLm9+L5X" \
-e SRC="vectorizer/roles" \
-e DEST="/tmp/minio-tests/examples/" \
jpweber/minio-init-dl:0.2.4
```

## kubernetes manifest additions

Include this in your current deployment

``` yaml
 initContainers:
  - name: data-loader
    image: jpweber/minio-init-dl
    env:
    - name: MINIO_URL
      value: "minio.k8sdev.example.com"
    - name: ACCESSKEY
      value: "AKIAIOSFODNN7EXAMPLE"
    - name: SECRETKEY
      value: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    - name: SRC
      value: "training-data/foo/"
    - name: DEST
      value: "/data"
    volumeMounts:
    - mountPath: /data
      name: db-storage
```
src vectorizer/native
dest /data01/vectorizer/native/

Assumes you already have a mount named `db-storage` in your main container. 

Create a secret to hold your access and secret keys