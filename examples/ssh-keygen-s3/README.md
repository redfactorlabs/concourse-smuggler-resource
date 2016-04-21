# Example: ssh-keygen-s3

Resource which generates a SSH key only once and uploads it to S3 by
wrapping the official S3 resource.

# How to use it?

## Pre-requisites

You need a S3 bucket and the right credentials

## Building the container

You must build the container image which would contain `smuggler`,
the `s3-resource` and tools like `jq` to modify json.

You need to create a docker hub image repository and run:
```
# Your container repository in docker-hub
export CONTAINER_TAG=<your image>

./build.sh
```

## Configure pipeline:

Now you can create and run the pipeline:

```
fly -t vagrant set-pipeline \
  -p ssh-keygen-s3 -c pipeline.yml \
  -v aws_access_key_id=$AWS_ACCESS_KEY_ID \
  -v aws_secret_access_key=$AWS_SECRET_ACCESS_KEY \
  -v bucket-name=$BUCKET_NAME
```
