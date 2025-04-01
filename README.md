lfs-minio
=========

Git LFS custom Transfer Type based on the Minio client to support S3 compatible buckets.

## Getting Started

Install:
* git-lfs (should be available in your package manager)
* lfs-minio `go install github.com/polyfloyd/lfs-minio@latest`

### Cloning an already configured repository

```sh
# Clone the repository as you would normally do.
git clone ...

cd <your repository>

# Enable LFS for a repository:
git lfs install

# Tell LFS to use lfs-minio as transfer program:
git config --add lfs.standalonetransferagent lfs-minio

# Set the path to the lfs-minio binary. If it is installed in your $PATH, this should work:
git config --add lfs.customtransfer.lfs-minio.path lfs-minio
```

lfs-minio reads environment variables to configure the remote. Set these, for example via direnv:
```
MINIO_BUCKET
MINIO_ENDPOINT
MINIO_ACCESS_KEY
MINIO_SECRET_KEY
```

> [!NOTE]
> MINIO_ENDPOINT does not expect a `https://` prefix!

Now you should be good to go! If this is an existing repository, pull in the missing objects in HEAD
with:
```sh
git lfs pull
```

Now add and commit the object files as you would with any other plain text files. It's also a good
idea to glance over the git-lfs manual page in just in case you ever need to do some plumbing.


# Configuring Credentials

## Scaleway

Navigate to [Organization -> API Keys](https://console.scaleway.com/iam/api-keys)

Use the purple `Generate an API key` to open the form:
* Select API key bearer: Myself (IAM user)
* Optionally provide a description to remind yourself what this key is used for.

Navigate to your bucket and go to the *Bucket Settings*. Under *Bucket information* you find the
*Bucket endpoint*. This bucket endpoint is the value of MINIO_ENDPOINT without the https:// and
bucket name prefix.

```sh
export MINIO_BUCKET=<name of the bucket>
export MINIO_ENDPOINT=s3.<region>.scw.cloud
export MINIO_ACCESS_KEY=<access key id>
export MINIO_SECRET_KEY=<secret key>
```

## Google Cloud Platform

Navigate to [Google Cloud Storage -> Settings -> Interoperability](https://console.cloud.google.com/storage/settings;tab=interoperability)

On the bottom of this page you find a section titled "Access keys for your user account".
Click "CREATE A KEY". This will give you an access key and secret that you need for the next step.

lfs-minio is configured via env vars, set the ones below via an .envrc,
```sh
export MINIO_BUCKET=<name of the bucket>
export MINIO_ENDPOINT=storage.googleapis.com
export MINIO_ACCESS_KEY=<access key>
export MINIO_SECRET_KEY=<secret>
```
