[![Build Status](https://travis-ci.org/redfactorlabs/concourse-smuggler-resource.svg?branch=master)](https://travis-ci.org/redfactorlabs/concourse-smuggler-resource)

# concourse-smuggler-resource: Concourse generic resource

Smuggler is a *"generic-resource"*, that allows you to quickly
implement any concourse resource with minimum boilerplate, even
in the pipeline itself.

It allows you to run any random command/script into the resource
container for the `check/in/out`.

Smuggler will basically:

 1. parse the input JSON from concourse
 2. set environment variables with the parameters
 3. execute the provided shell script
 4. parse the output and create a valid response for concourse.

*Smuggler* is ideal for PoC, prototyping, fast development or implementation
of simple resources.

## Basic example

The following resource definition will generate random numbers using
the [sh `${RANDOM}` variable](http://tldp.org/LDP/abs/html/randomvar.html).

`get` in the job `some_job` writes the random number in the file specified
by the parameter `params.target_file`. It also provides the current date in
the metadata.

```
# Add the custom resource type
# See https://concourse.ci/configuring-resource-types.html
resource_types:
- name: smuggler
  type: docker-image
  source:
    repository: redfactorlabs/concourse-smuggler-resource
    tag: alpine

resources:
# A randon number generator
- name: random_number_generator
  type: smuggler
  source:
    smuggler_debug: true
    target_file: random_number_file
    commands:
      check: |
        echo "${RANDOM}" > ${SMUGGLER_OUTPUT_DIR}/versions
      in: |
        echo "${SMUGGLER_VERSION_ID}" > \
          ${SMUGGLER_DESTINATION_DIR}/${SMUGGLER_target_file}
        echo "date=$(date)" > ${SMUGGLER_OUTPUT_DIR}/metadata

jobs:
# A example job using our new resource
- name: some_job
  plan:
  - get: random_number_generator
    trigger: true
    params:
      target_file: my_random_number_file
  - task: print_the_number
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: alpine
      inputs:
      - name: random_number_generator
      run:
        path: sh
        args:
        - -ec
        - |
          cat random_number_generator/my_random_number_file
```

# Other examples

Check [the examples directory](https://github.com/redfactorlabs/concourse-smuggler-resource/tree/master/examples)
for examples of hacks and resources.

Some to highlight:

 * [ssh-keygen resource](https://github.com/redfactorlabs/concourse-smuggler-resource/blob/master/examples/ssh-keygen/pipeline.yml)
 * [ssh-keygen-s3 resource](https://github.com/redfactorlabs/concourse-smuggler-resource/tree/master/examples/ssh-keygen-s3): Generates a SSH key and stores it in S3.
 * [s3-with-default resource](https://github.com/redfactorlabs/concourse-smuggler-resource/tree/master/examples/s3-with-default): Extends the official S3 resource to allow define a default value if the object is missing.

# Using smuggler-concourse

## Defining `check/in/out` commands

Use `commands.check`, `commands.in` or `commands.out` to
provide the script to execute. All of them are optional:

 * `commands.check`: Called to check and find new versions.
    Write the versions to `${SMUGGLER_OUTPUT_DIR}/versions`, one line for each version.

 * `commands.in`: called by the [`get` step](https://concourse.ci/get-step.html)
    to fetch a resource. Write the data into `${SMUGGLER_DESTINATION_DIR}`

 * `commands.out`: called by the [`put` step](https://concourse.ci/put-step.html)
    to upload the resource.
    The input files from previous steps in the job would be in `${SMUGGLER_SOURCES_DIR}`

## Input & output

The  `check/in/out` scripts communicate with smuggler/concourse via:

 * Environment variables:

   | Variable                   | example               | available in   | description |
   |----------------------------|-----------------------|----------------|-------------|
   | `SMUGGLER_<param_name>`    | `SMUGGLER_id_rsa`     | `check/in/out` | Parameters from `source.*` or `params.*` |
   | `SMUGGLER_VERSION_<key>`   | `SMUGGLER_VERSION_ID` | `check/in`     | Environment variable with the latest resource version retrieved. \\ Not be defined in first run of `check`. |
   | `SMUGGLER_OUTPUT_DIR`      |                       | `check/in/out` | The directory to write versions and metadata. |
   | `SMUGGLER_DESTINATION_DIR` |                       | `in`           | The directory to write the retrieved data to. |
   | `SMUGGLER_SOURCES_DIR`     |                       | `out`          | The directory with files from previous steps in the job |

   > **Important**: Note that `SMUGGLER_OUTPUT_DIR` with
   > `SMUGGLER_DESTINATION_DIR` or `SMUGGLER_SOURCES_DIR` are
   > different directories.

 * `${SMUGGLER_OUTPUT_DIR}/versions`: For `check/in/out`.
   * **Optional**, only processed if no json is written in `stdout`.
   * Smuggler will automatically  add the default key `ID`.
   * Restrictions:
     * `check`: Your command **must** write here the versions found, one line per version.
     * `in`: Optional, if no version is written, smuggler will use the same as
       passed to the command as input.
       Only the first line is taken into account.
     * `out`: *Mandatory*, you must always specify a version for out, as
       concourse does not provide the version in the input.
       Only the first line is taken into account.

 * `${SMUGGLER_OUTPUT_DIR}/metadata`: For `in/out` *Optional.* the
   metadata for concourse as a multiline file with `key=value` pairs.

 * `${SMUGGLER_DESTINATION_DIR}/`: For `in`.
   The directory to write the retrieved data to.

 * `${SMUGGLER_SOURCES_DIR}/*/*`: For `out`. The directories with files from previous steps in the job.

 * `stdin`: For `check/in/out`. Raw JSON with as it is
   sent from concourse and [as described in the implementing concourse resources documentation.](https://concourse.ci/implementing-resources.html)

   This allows your command parse the request directly, or pass it to a
   wrapped resource.

   > **Note**: If `filter_raw_request: true`, all the specific smuggler
   > configuration will be filtered out (`source.commands`,
   > `source.smuggler_params`, `params.smuggler_params`, etc.).

 * `stdout`: For `check/in/out`, **Optional**. verbatim JSON response
   request [as described in the implementing concourse resources documentation.](https://concourse.ci/implementing-resources.html)

   > **Note**: if you print anything to stdout that is not JSON, the output
   > will be not be passed to concourse, but instead dump to `stderr`.

## Resource parameters

Any additional parameter in the `source` of the definition,
or passed to the `get` or `put` step as params, would be passed as environment
variables to the script with the prefix `SMUGGLER_`.
e.g.`SMUGGLER_param1=value1`, `SMUGGLER_param2=value2`.

For example in this definition:

```
resources:
- name: random_number_generator
  type: smuggler
  source:
    commands:
      check: |
        ...
      in: |
        ...
    global_config_entry: value1

jobs:
- name: some_job
  plan:
  - get: random_number_generator
    trigger: true
    params:
      specific_get_config_entry: value2
```

Smuggler would set `SMUGGLER_global_config_entry` for `check` and `in`, and
`SMUGGLER_specific_get_config_entry` for the `in` command.

## Smuggler specific parameters

Smuggler understands these parameters:

 * `commands.{check,in,out}` to define the commands as described above.

 * `smuggler_debug: [true|false]`. *Optional*. it will print debugging
   information to the `stderr`.

 * `filter_raw_request: [true|false]`: *Optional*. Would remove the
   smuggler specific parameters from the JSON passed via `stdin` to
   the script.

 * `smuggler_params.<param>`: *Optional*. Allows group the parameters so they can filtered
   out with `filter_raw_request`.

## Parameter priorities

Parameters can be defined in different places so parameters
with the same name would be overridden depending where they are declared
(first has more priority)

 1. `/opt/resource/smuggler.yml` in the docker image.
 1. resource definition, `source.smuggler_params.<param>`
 1. resource definition, `source.<param>`
 1. `get/put` step, `params.smuggler_params.<param>`
 1. `get/put` step, `params.<param>`

This allows easily define default values for parameters in your resources.

## Logging and troubleshooting

All the operations would log into `/tmp/smuggler.log` in the container. Use
the parameter `smuggler_debug: true` to print the log to `stderr`
that would display the log in the concourse UI.

You can [intercept the container](https://concourse.ci/fly-intercept.html)
to read the log or interact with the commands directly:

```
fly -t demo intercept -c pipeline_name/resource_name # intercept a check

fly -t demo intercept -j pipeline_name/job_nome # intercept a get/put
```

In `/tmp/smuggler.log` you can find the exact command used to call the resource,
so you can execute it again by copy&paste for quick troubleshooting:

```
2017/09/27 23:23:32 [INFO] Smuggler command called as:
/opt/resource/in /tmp/build/get <<"EOF"
{
  "source": {
    "commands": {
      "check": "echo \"${RANDOM}\" \u003e ${SMUGGLER_OUTPUT_DIR}/versions\n",
      "in": "echo \"${SMUGGLER_VERSION_ID}\" \u003e ${SMUGGLER_DESTINATION_DIR}/random_number\n"
    },
    "smuggler_debug": true
  },
  "version": {
    "ID": "1480"
  }
}
EOF
```

# Advanced usage

## Bundle smuggler configuration into the docker image

You can optionally write the same configuration of the `source` section in
the resource container image, in `/opt/resource/smuggler.yml`.

The content of that file will be merged with the request, so that any parameter
and command defined in the pipeline, will override the ones defined in
`smuggler.yml`.

This way smuggler becomes a framework to create any kind of resource with
very little boilerplate.

## Wrapping other resources with smuggler

Smuggler passes the raw JSON request from concourse from `stdin` and
returns back the response from `stdout` (if it is a valid response).

Additionally, with `source.filter_raw_request` all the smuggler config
will be strip from the request in `stdin`.

This way it is really easy to wrap any third party resource and
change their behaviour. Simply copy the other resource commands in your
image and call them directly.

For example, use S3 resource to generate some default content if the file is
not in the bucket, and behave as usual otherwise:

```
---
filter_raw_request: true
commands:
  check: |
    /opt/resource/wrapped/s3/check > response.json
    # If it is the first run, just dispatch a - string to for 'in' to be triggered
    jq 'if . == [] then [{ "version_id":"-"}] else . end' < response.json

  in: |
    if [ "${SMUGGLER_VERSION_version_id}" == "-" ]; then
      # First run, simply print the default content
      echo "${SMUGGLER_default_content}" > ${SMUGGLER_DESTINATION_DIR}/${SMUGGLER_versioned_file}
    else
      # First run, simply print the default content
      /opt/resource/wrapped/s3/in ${SMUGGLER_DESTINATION_DIR}
    fi

  out: /opt/resource/wrapped/s3/out ${SMUGGLER_SOURCES_DIR}
```

## Complex commands and inline scripts

Commands can be defined using these two syntaxes:

 1. a `bash`/`sh` script using [multiline literal strings in yaml](http://www.yaml.org/spec/1.2/spec.html#id2795688)

    This is great for simple bash scripts.

 2. A hash with `path: <string>` and `args: [<string>, ...]`

    This would allow you to use any embedded scripting language in your
    definition, like `bash`, `python`, `perl`, `ruby`...


## Supported tags and Dockerfiles

 * `alpine` or `x.x.x-alpine` [Dockerfile.alpine](https://github.com/redfactorlabs/concourse-smuggler-resource/blob/master/Dockerfile.alpine)
 * `ubuntu` or `x.x.x-ubuntu` [Dockerfile.ubuntu](https://github.com/redfactorlabs/concourse-smuggler-resource/blob/master/Dockerfile.ubuntu)

# Contributions

Smuggling is fun! Share it! Send over or comment us your hacks and implementations.

See the [AUTHORS](AUTHORS.md) file for contributions.

# Credits

I stoled a lot of code around in github, specially from other resources
like `s3-resource`. Thanks to all of you!
