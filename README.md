[![Build Status](https://travis-ci.org/redfactorlabs/concourse-smuggler-resource.svg?branch=master)](https://travis-ci.org/redfactorlabs/concourse-smuggler-resource)

# concourse-smuggler-resource

Concourse generic resource, to quickly implement any kind of resource by
defining any command for the `check`, `get` and `put` actions.

*Smuggler* is ideal for PoC, prototyping, fast development or implementation
of simple resources based on existing command line tools, hacking or extending
existing resources, etc.

## Resource definition

> Note: It is recommended that you have a look at how [custom resources are implemented](https://concourse.ci/implementing-resources.html)

You can easily register smuggler as a service by using
[custom resource type definitions](https://concourse.ci/configuring-resource-types.html):

First build the container:

```
# A repository in dockerhub or other docker repository
export SMUGGLER_DOCKER_TAG=<...your_docker_repository...>
./scripts/build-docker
```

> We will eventually distribute some prebuilt images, but right now
> is better you build your own.

```
resource_types:
- name: smuggler
  type: docker-image
  source:
    # Replace the url
    repository: <...your_docker_repository...>
```

Modify `Dockerfile` to build your own container image bundled with smuggler.

As it is a unique static compiled binary it should work on any distribution
with any tools and scripts you need to create your custom resource.

See below for more details.

## Source configuration and tasks

Once you `smuggler` is defined as a resource type, you only need to define
your resource using this structure:

```
resources:
- name: my_smuggler_resource
  type: smuggler
  source:
    commands:
      check:
        path: <command>
        args:
        - ...
      in:
        path: <command>
        args:
        - ...
      out: "<command> <args>"

    filter_raw_request: true

    smuggler_params:
    - key1: value1
    - key2: value2
    - ...

    # Additional random non smuggler source parameters
    other_param_key1: value1
    other_param_key2: value2
    ...


jobs:
- name: some_job
  plan:

  - get: my_smuggler_resource
    params:
      smuggler_params:
        # Override the existing key
        key1: other_value1
      # Additional random non smuggler source parameters
      key3: value3

  - put: my_smuggler_resource
    params:
      smuggler_params:
        # Override the existing key
        key1: other_value1
      # Additional random non smuggler source parameters
      key3: value3

```

The `source` configuraton includes:

 * `commands`: definition of the
   commands `check/in/out` called from concourse to `check` new
   versions and `get` or `put` resources.

   It can be defined as:

    * Single string: Each one gets the command in one string which will
      shell out to `bash -o -u -o pipefail -c <cmd>` or `sh -o -u -c <cmd>`.
      If none of `bash` or `sh` is found, the command will be executed directly.

    * as a hash of `path` and `args` similar to [concourse task
      `run` definition](https://concourse.ci/running-tasks.html#run)

> All commands are *optional*, and if not defined they will execute a
> dummy operation (Of course you always want to define at least one ;)).

 * `filter_raw_request`: Filter the smuggler specific config from the
   verbatin json request passed via `stdin` to the commands.

 * `smuggler_params`: **Optional**. List of key-value pairs to pass to
   all the commands.

   All these parameters will be passed as environment variables prefixed with
   `SMUGGLER_`: `SMUGGLER_key1=value1`, `SMUGGLER_key2=value2`

 * Any additional parameter in the source, like `other_param_key1` or
   `other_param_key1` will be threated as `source.smuggler_params`, but will not
   be filtered when `filter_raw_request` is enabled.

   These keys will override any value in `source.smuggler_params` with the same key.

The `get/put` task configuration includes:

 * `smuggler_params`: **Optional**. Similar to `source.smuggler_params`,
   list of key-value pairs to pass to the command as environment variables
   prefixed with `SMUGGLER_`.

   They will have precedence any parameter in `source`.

 * Any additional parameter in the source, will be threated as
   `params.smuggler_params`, but will not be filtered when `filter_raw_request`
   is enabled.


## Behavior

You can use any of the tasks related to this resource: `check`, `get` and `put`.

## Shared input for `check`, `in` and `out`

Each command will get some input via environment variables or `stdin`:

 * `SMUGGLER_<param_name>`: For `check/in/out`. Environment variables with the
   prefixed source parameters under `source.smuggler_params` or directly
   `source`.

   For `in/out` will also include the parameters under `params.smuggler_params`
   or `params`

 * `SMUGGLER_VERSION_<key>`: For `check/in` Environment variable with the latest
   resource version, one version per key-value pair in the version.
   They will not be defined if it is first run of `check`.

 * `SMUGGLER_VERSION_ID`: Default version variable when using plain string
   versions (see below for output details).

 * `SMUGGLER_OUTPUT_DIR`: For `check/in/out`. The directory path to write the
   resulting versions and metadata when not using `stdout`.

 * `SMUGGLER_DESTINATION_DIR`: For `in`. The directory path to write the data to.

 * `SMUGGLER_SOURCES_DIR`: For `out`. The directory path with the
   build's full set of sources.

> **Important**: do not mix up `SMUGGLER_OUTPUT_DIR` with
> `SMUGGLER_DESTINATION_DIR` or `SMUGGLER_SOURCES_DIR`

 * `stdin`: For `check/in/out`. Verbatin json with all the structure as is
   sent from concourse. This allows your command parse the request directly.

   If `source.filter_raw_request` is `true`, all the specific smuggler
   configuration will be filtered out (`source.commands`,
   `source.smuggler_params`, `params.smuggler_params`, etc.). This is useful
   when wrapping third party resources (see below).

Output to send to concourse to the commands:

 * `stdout`: For `check/in/out`,  **Optional**. the verbatin JSON response
    request [as described in the implementing concourse resources documentation.]
    (https://concourse.ci/implementing-resources.html)

> **Note** if you print to stdout and it is not JSON, the output will be discarded.

 * `${SMUGGLER_OUTPUT_DIR}/versions`: For `check/in`.
   * **Optional**, only processed if no json is written in `stdout`.
   * Smuggler will add the default key `ID`.
   * For `check`: Your command **must** write here the versions found, one line per version.
   * For `in`: If no version is written, smuggler will use the same as
     passed to the command. Only the first line will be taken into account.

 * `${SMUGGLER_OUTPUT_DIR}/metadata`: For `in/out` *Optional.* the
   metadata for concourse as a multiline file with `key=value` pairs
   separated by `=`.

 * `${SMUGGLER_OUTPUT_DIR}/versions`: For `check/in`.
   * **Optional**, only processed if no json is written in `stdout`.
   * For `check`: Your command **must** write here the versions found, one line per version.
   * For `in`: If no version is written, smuggler will use the same as
     passed to the command. Only the first line will be taken into account.
   * If each line is a valid JSON, they will be interpreted.

### `check` Find out what you want to smuggle

Will execute the command configured as `check`.

### `get` and `put` smuggle into and out concourse

Will execute the commands configured as `in` and `out` respectively.

## Complex commands and inline scripts

You can smuggle even more if you use inline scripts included as
[multiline literal strings in yaml](http://www.yaml.org/spec/1.2/spec.html#id2795688)
in your command definition:

```
check:
  path: sh
  args:
  # sh reads commands from next argument with -c
  - -c
  # all the script goes here
  - |
    echo "this is"
    echo "a multiline script \o/"
```

This way you pass almost any embedded script language in your scripts, like
`bash`, `python`, `perl`, `ruby`...

For example:

 * `bash/sh` with `-c` option:
   ```
resources:
- name: generate-ssh-key
  type: smuggler
  source:
    commands:
      out:
        path: sh
        path: <command>
        args:
        - -e
        - -c
        - |
          ssh-keygen -f id_rsa -N ''
          tar -cvzf $SMUGGLER_DESTINATION_DIR/id_rsa.tar.gz id_rsa id_rsa.tgz
```

# Advanced usage

## Bundle smuggler configuration in `/opt/resource/smuggler.yml` in resource image

You can optionally write the same configuration of the `source` section in
the resource container image, in `/opt/resource/smuggler.yml`.

The content of that file will be merged with the request, so that any parameter
and command defined in the pipeline, will override the ones defined in
`smuggler.yml`.

This enables you to encapsulate all the implementation and not expose
it in the pipelines, keeping the pipeline simple and letting you to
distribute and reuse the resource as desired.

## Wrapping other resources with smuggler

As mention, you can read the raw JSON request from concourse from `stdin`, and write
it directly the response to `stdout`. Additionally, with `source.filter_raw_request`
all the smuggler config will be removed from the resquest.

With these features, it is really easy to wrap any third party resource and
change their behaviour, you only need to bundle the resource in your image
and shell out the resource.

For example, to use S3 to store generated keys with `ssh-keygen`:

```
- name: ssh_key_on_s3
  type: smuggler-s3
  source:
    bucket: my-ssh-keys
    versioned_file: id_rsa.tgz
    access_key_id: ACCESS-KEY
    secret_access_key: SECRET

    commands:
      check:
        path: bash
        args:
        - -c
        - -e
        - |
          if [ -z "$SMUGGLER_VERSION_ID" ] &&
            # First time we will generate the key
            echo "initial" > ${SMUGGLER_OUTPUT_DIR}/versions
          else
            /opt/resource/wrapped_resource/s3/check
          fi
      in:
        path: bash
        args:
        - -c
        - -e
        - |
          if [ "$SMUGGLER_VERSION_ID" == "initial" ] &&
            # First time we will generate the key
            ssh-keygen -f id_rsa -N ''
            tar -czf ${SMUGGLER_DESTINATION_DIR}/${SMUGGLER_versioned_file} id_rsa id_rsa.pub

            # And the upload the key with an out command.
            #
            # The out command will read the "in" request from stdin and
            # write a response to stdout both they are compatible
            #
            /opt/resource/wrapped_resource/s3/out ${SMUGGLER_SOURCES_DIR}
          else
            /opt/resource/wrapped_resource/s3/in ${SMUGGLER_DESTINATION_DIR}
          fi
      out:
        path: bash
        args:
        - -c
        - -e
        - |
          /opt/resource/wrapped_resource/s3/out ${SMUGGLER_SOURCES_DIR}

```

# Examples

Check [the examples directory](https://github.com/redfactorlabs/concourse-smuggler-resource/tree/master/examples)
for examples of hacks and resources.

# Contributions

Smuggling is fun! Share it! Send over or comment us your hacks and implementations.

# Credits

I stoled a lot of code around in github, specially from other resources
like `s3-resource`. Thanks to all of you!
