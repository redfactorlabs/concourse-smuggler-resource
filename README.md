[![Build Status](https://travis-ci.org/redfactorlabs/concourse-smuggler-resource.svg?branch=master)](https://travis-ci.org/redfactorlabs/concourse-smuggler-resource)

# concourse-smuggler-resource

Concourse generic resource, to quickly implement any kind of resource by
defining any command for the `check`, `get` and `put` actions.
i
*Smuggler* is ideal for PoC, prototyping, fast development or implementation
of simple resources based on existing command line tools.

## Resource definition

> Note: It is recommended that you have a look at how [custom resources are implemented](https://concourse.ci/implementing-resources.html)

You can easily register smuggler as a service by using
[custom resource type definitions](https://concourse.ci/configuring-resource-types.html):

```
resource_types:
- name: smuggler
  type: docker-image
  source:
    repository: redfactorlabs/concourse-smuggler-resource#ubuntu-14.04
```

Alternatively, you can build your own container image bundled with smuggler,
which is a static compiled binary, and any tools and script you need to
create your custom resource. See below for more details.

## Source configuration

Once you `smuggler` is defined as a resource type, you only need to define
your resource using this structure:

```
resources:
- name: <resource-name>
  type: smuggler
  source:
    commands:
    - name: check
      path: <command>
      args:
      - ...
    - name: in
      path: <command>
      args:
      - ...
    - name: out
      path: <command>
      args:
      - ...
    extra_params:
    - key1: value1
    - key2: value2
    - ...
```

The `source` configuraton includes:

 * `commands`: *Optional*. Each command definition for `check/in/out` commands
   called from concourse to `check` new versions and `get` or `put` resources.
   Each command has a `path` and `args` similar to
   [concourse task `run` definition](https://concourse.ci/running-tasks.html#run)

   All commands are *optional*, and if not defined they will execute a
   dummy operation (Of course you always want to define at least one ;)).


 * `extra_params`: **Optional**. List of key-value pairs to pass to
   all the commands.

   All these parameters will be passed as environment variables prefixed with
   `SMUGGLER_`: `SMUGGLER_key1=value1`, `SMUGGLER_key2=value2`

## Behavior

You can use any of the tasks related to this resource: `check`, `get` and `put`.

### `check` Find out what you want to smuggle

Will execute the command configured as `check`.

Input of the script:

 * `SMUGGLER_<source_extra_param_name>`: Environment variables with the
   prefixed source parameters defined in `extra_params`.
 * `SMUGGLER_VERSION_ID`: Environment variable with the latest resource
   version. It will be a empty string in the first run.
 * `SMUGGLER_OUTPUT_DIR`: The directory path to write the resulting versions.

Output to send to concourse:
 * `${SMUGGLER_OUTPUT_DIR}/versions`: Your command **must** write here the
   versions found, one line per version.

   If the file is not created, `check` will error.

### `get` and `put` smuggle into and out concourse

Will execute the commands configured as `in` and `out` respectively.

Input of the script:

 * `SMUGGLER_<param_name>`: Environment variables with the
   specific parameters passed to `get` or `put`.
 * `SMUGGLER_<source_extra_param_name>`: Environment variables with the
   prefixed source parameters defined in `extra_params`.
 * `SMUGGLER_OUTPUT_DIR`: The directory path to write the resulting version
   and metadata.

 * `SMUGGLER_DESTINATION_DIR`: *Only `get`*. The directory path to write the data to.
 * `SMUGGLER_SOURCES_DIR`: *Only `put`*. The directory path with the
   build's full set of sources.

> **Important**: do not mix up `SMUGGLER_OUTPUT_DIR` with
> `SMUGGLER_DESTINATION_DIR` or `SMUGGLER_SOURCES_DIR`

Output to send to concourse.

 * `${SMUGGLER_DESTINATION_DIR}/*`: *Only `get`*. The retrieved data.

 * `${SMUGGLER_OUTPUT_DIR}/versions`: *Only `get`, Optional.* the version retrieved.
   Only the first line will be used. (Note, it is `versions` not `version`)
 * `${SMUGGLER_OUTPUT_DIR}/metadata`: *Optional.* the metadata for concourse as
   a multiline file with `key=value` pairs separated by `=`


## Complex commands and inline scripts

You can smuggle even more if you use inline scripts included as
[multiline literal strings in yaml](http://www.yaml.org/spec/1.2/spec.html#id2795688)
in your command definition:

```
- name: check
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
    - name: out
      path: sh
      path: <command>
      args:
      - -e
      - -c
      - |
        ssh-keygen -f id_rsa -N ''
        tar -cvzf $SMUGGLER_DESTINATION_DIR/id_rsa.tar.gz id_rsa id_rsa.tgz
```
 * `python`: TODO
   ```
python -c '
friends = ["john", "pat", "gary", "michael"]
for i, name in enumerate(friends):
    print "iteration {iteration} is {name}".format(iteration=i, name=name)
'
```
 * `ruby`: TODO

# Advanced usage

## Smuggler as framework for new resources

TODO ... explain config.yml

## Examples

TODO

## Contributions

Smuggling is fun, share it! Send over or comment us your hacks and implementations.

## Credits

I stoled a lot of code around in github, specially from other resources
like `s3-resource`. Thanks to all of you!
