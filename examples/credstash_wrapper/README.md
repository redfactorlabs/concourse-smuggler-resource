# Example: credstash wrapper

Proof of concept to add credential management to resources in concourse.

This resource uses [smuggler concourse resource](https://github.com/redfactorlabs/concourse-smuggler-resource)
to allow any other resource
to retrieve variables from a [credstash store](https://github.com/fugue/credstash)

It relies on [unicreds](https://github.com/Versent/unicreds) (golang credstash implementation)
to retrieve all the variables from a credstash store, and [spruce](https://github.com/geofffranks/spruce)
to expand the variables using the `(( grab $credential.name ))`.

We use a basic shell script and static golang binaries, so this resource can be use for probably any resource.

## Future work and limitations

 * I will create a custom `docker_image` resource able to automatically path any docker image with a resource.

 * Currently it uses [spruce and the spruce syntax](https://github.com/geofffranks/spruce), but in the future we might try
   to implement a wrapper adhoc that simulates [the proposal from concourse devs](https://github.com/concourse/concourse/issues/291#issuecomment-233194564): `${credstash/a.random.key}`

 * If you use `spruce` or `spiff` to render your manifest, you might have issues with this :)

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

Define the new resource types:

```
resource_types:
- name: smuggler-credstash-git
  type: docker-image
  source:
    repository: redfactorlabs/smuggler-credstash-git-resource

- name: smuggler-credstash-s3
  type: docker-image
  source:
    repository: redfactorlabs/smuggler-credstash-s3-resource
```

> **NOTE:** If you are going to use this resources **I highly recommend** push your
> own copy to your own docker repository. This image might change.

And then define the resource, with these two kind of parameters:

 1. Credstash configuration:
    * `credstash_table`: optional, credstash store dynamo table
    * `credstash_region`: AWS region of the credstash store. Default: `eu-west-1`
    * `credstash_aws_access_key_id`, `credstash_aws_secret_access_key` and `credstash_aws_session_token`: AWS credentials. Not needed when using IAM instance profiles
    * `credstash_aws_iam_profile`: optional, use IAM instance profile (see below)

 2. The normal parameters of the resource, using a [syntax like in spruce](https://github.com/geofffranks/spruce).
    For example: ```private_key: "(( grab $github.id_rsa ))"```

For example:

```yaml
- name: project-git-iam-profile
  type: smuggler-credstash-git
  source:
    credstash_aws_access_key_id: {{credstash_aws_access_key_id}}
    credstash_aws_secret_access_key: {{credstash_aws_secret_access_key}}
    credstash_aws_session_token: {{credstash_aws_session_token}}
    credstash_table: concourse-demo
    uri: git@github.com:/concourse-demo
    private_key: "(( grab $github.id_rsa ))"

```

## Using credstash with IAM instance profiles

One way of using this resource is [granting your Concourse workers with a IAM profile](https://bosh.io/docs/aws-iam-instance-profiles.html)
that would allow access the credstash without having to pass any credentials at all to the pipelines.

In this branch there is an example of how to configure the store and the profile: https://github.com/keymon/tf-credstash/tree/implement_iam_profile

You only need to pass the option  `aws_iam_profile: true` as to the resource:

```
- name: project-git-iam-profile
  type: smuggler-credstash-git
  source:
    aws_iam_profile: true
    credstash_table: credstash-concourse-demo
    uri: git@github.com:/concourse-demo
    private_key: "(( grab $github.id_rsa ))"
```

## How does it work?

This resources basically "intercepts" [the json request](https://concourse.ci/implementing-resources.html), and expands the variables in it with values from credstash.

The implementation can be checked in:

 * `Dockerfile.git-resource`: It gets the official image, renames the `check`, `in`, `out`, adds the binaries for `spruce`, `smuggler` and `unicreds` and the scripts.
 * `smuggler.yml`: uses smuggler to simply delegate to the `wrapper.sh` for each command.
 * `wrapper.sh`: Implements all the logic:
  1. The json request from the stdin
  2. Calls `unicreds exec` to retrieve the credentials, set them as environment variables and call `spruce`
  3. `spruce merge`: would expand the environment variables of the json from stdin
  4. `spruce json`: spruce reads json and spits yaml. This conversts back to json.


