# Example: s3-with-default

Simple resource which allows define a S3 file, with some default
value in case the file does not exist in the bucket.

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
  -p s3-with-default -c pipeline.yml \
  -v aws_access_key_id=$AWS_ACCESS_KEY_ID \
  -v aws_secret_access_key=$AWS_SECRET_ACCESS_KEY \
  -v bucket-name=$BUCKET_NAME
```
