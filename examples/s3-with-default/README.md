# Example: s3-with-default

Simple resource which allows define a S3 file, with some default
content or a default command to run in case the file does not exist
in the bucket and it is the first time to run.

This solves the initialisation issue caused by the official S3 resource
failing when running get and the file is missing.

# How to use it?

## Parameters

You can specify the following optional parameters:

 * `default_content`: Default text to populate the file with.
 * `default_command`: Command to run to generate the file.

If none is set, no content will be generated.

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
