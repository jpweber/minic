# Minic

Download data from minio via an init container. Useful pre-loading volumes for stateful applications, or anything that needs data before the applications starts. 


## Parameters
Parameters are read from the following environment variables  
- `MINIO_URL` = address and port for the minio server. Does not need scheme.  
- `ACCESSKEY` = Minio access key  
- `SECRETKEY` = Minio secret key  
- `SRC` = Source of the files you wish to copy to your container.  
Takes form of `bucketname/path/file` or `bucketname/path` for all contents in a folder  
- `DEST` = Path to where you wish data to inside your container  

Special note for the `DEST` paramert. If you are downloading all files from a dir, don't include that dir name in your `DEST` Value. For example. If you were downloading all the contents from `users/roles` do _not_ make your destination `/data/users/roles`. *Do* make it `/data/users`. The roles directory will be created as part of the download. 



## Using it as init container in Kubernetes

Include this in your current deployment. Modify the `env` values to match your environment.  The name of the container can be changed to whatever you wish. Below example asssumes you already have a mount named `db-storage` in your main container. This should match where your primary container is going to read its data. 

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

TODO:
Create a secret to hold your access and secret keys

## Run locally for testing

### Binary CLI
Set your variables like the example below
``` shell
export MINIO_URL="192.168.2.66:9000"
export SRC="vectorizer/roles"
export ACCESSKEY="LEB3JJ3OCBN4HTDIS5IZ"
export SECRETKEY="xs3fx83cMkV7Oh+6jlGGTt9kTmT5D6yoQLm9+L5X"
export DEST="/tmp/minio-tests/examples"
```
Then you can run the binary with out any extra parameters

### Docker Container

``` shell
docker run -ti -v /tmp:/tmp \
-e MINIO_URL="192.168.2.66:9000" \
-e ACCESSKEY="LEB3JJ3OCBN4HTDIS5IZ" \
-e SECRETKEY="xs3fx83cMkV7Oh+6jlGGTt9kTmT5D6yoQLm9+L5X" \
-e SRC="vectorizer/roles" \
-e DEST="/tmp/minio-tests/examples/" \
jpweber/minio-init-dl:0.2.4
```