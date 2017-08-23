# Example: credstash resource

Proof of concept to get credentials from credstash

This resource uses [smuggler concourse resource](https://github.com/redfactorlabs/concourse-smuggler-resource)
to get credentials from [credstash store](https://github.com/fugue/credstash)

Inspired on the code from https://hub.docker.com/r/paroxp/concourse-credstash-resource/


## Usage example

You must first configure a [credstash store](https://github.com/fugue/credstash)
using credstash, unicreds, or a [terraform configuration](https://github.com/keymon/tf-credstash).

You must then populate the credstash store with variables:

```
unicreds \
  -r eu-west-1 \
  -t credstash-concourse-demo \
  put-file \
  -k alias/credstash-concourse-demo \
  github.id_rsa \
  ~/.ssh/concourse-demo
```

Define the new resource type:

```
resource_types:
- name: smuggler-credstash
  type: docker-image
  source:
    repository: redfactorlabs/smuggler-credstash-resource
```

> **NOTE:** If you are going to use this resources **I highly recommend** push your
> own copy to your own docker repository. This image might change.

And then define the resource these parameters:

  * `credstash_table`: optional, credstash store dynamo table
  * `credstash_region`: AWS region of the credstash store. Default: `eu-west-1`
  * `credstash_aws_access_key_id`, `credstash_aws_secret_access_key` and `credstash_aws_session_token`: AWS credentials. Not needed when using IAM instance profiles
  * `credstash_aws_iam_profile`: optional, use IAM instance profile (see below)

For example:

```yaml
- name: credstash-aws-creds
  type: smuggler-credstash
  source:
    credstash_table: credstash-concourse-demo
    credstash_key: alias/credstash-concourse-demo
    credstash_aws_access_key_id: {{credstash_aws_access_key_id}}
    credstash_aws_secret_access_key: {{credstash_aws_secret_access_key}}
    credstash_aws_session_token: {{credstash_aws_session_token}}
```
### Get secrets

Finally you can consume the credentials by passing a parameters `secrets` with the list of secrets to query:

```
- get: credstash-aws-creds
  params:
    secrets:
    - example.key.date
    - deploy.test-server.TOKEN
    - github.id_rsa

```


The secrets will be stored in files named as the key value, eg `./credstash-aws-creds/deploy.test-server.TOKEN`

### Set secrets

In order to set the value of secrets, you can create a task that populates files named as the key of the secret and with the value as content:

```
- put: credstash-aws-creds
  params:
    secrets_dir: new_secrets
```

## Using credstash with IAM instance profiles

One way of using this resource is [granting your Concourse workers with a IAM profile](https://bosh.io/docs/aws-iam-instance-profiles.html)
that would allow access the credstash without having to pass any credentials at all to the pipelines.

In this branch there is an example of how to configure the store and the profile: https://github.com/keymon/tf-credstash/tree/implement_iam_profile

You only need to pass the option  `aws_iam_profile: true` as to the resource:

```
- name: credstash-iam-profile
  type: smuggler-credstash
  source:
    credstash_table: credstash-concourse-demo
    credstash_key: alias/credstash-concourse-demo
    credstash_aws_iam_profile: true
```


