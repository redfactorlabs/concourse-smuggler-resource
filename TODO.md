# Brainstorming

Generic resource which will allow to specify in Bash/shell any actions for
check, in and out in the manifest itself.

My intended implementation and ideas are:

  * Implemented in Golang and built statically:
    * Can be installed in any container, so the user can create new resources
      with the tools they need.
    * Will support specify the "script" to run for check/in/out in shell script
      by passing some configuration in the resource definition.
    * The configuration would allow define additional parameters to read in
      the resource definition (for check) or for get/put actions.
    * The configuration can be passed also as a file in the container of the
      resources (e.g. `/opt/resource/config.yml`). This would allow create
      a functional resource ready to be shared as docker containers.
    * Will provide common backends to use:
      * Desired: S3 backend, to store/download files automatically in S3.
    * Optionally: Allow define a custom resource with custom resource type
      from any docker container. It would basically generate the resource
      content by:
       * Downloading a specified docker image
       * Adding the `check/in/out` binaries from `smuggler-resource`
       * Add the `/opt/resource/config.yml`

## Idea for wrapping commands:

Do not implement the wrapped resources as described in #1,
but let the users to wrap the  wrapped command directly.

Make it easy to do by:

 * Pass and read request/response from stdin/stdout
 * add an option to filter out the "smuggler" specific options or not.
 * some smuggler params can be under a key, so they can be filtered out. The
   smuggler commands read the aprams from envvars
 * document how to close/redirect stdin/stdout in bash to avoid poluting the
   output and input.

This way, the user can simply call the other resource directly, which
will read and write the data.

If the user needs to change the response or request, that is an advance
use case and they should implement something by themselves. We can include
some basic examples using jq.

## Add a quick shell out option

Add a quick shell out config for check/in/out, which will:
 * first try to run the line in bash
 * second in sh
 * finally, run it directly

like:

```
source:
  smuggler_config:
    check: "ls something/* | shasum | cut -f 1 -d ' '"
    in: "cp something/* dest/*"
```

# MVP

 * [X] Basic check/in/out which runs a script command
 * [X] Add integration tests of the final commands
 * [X] Add travis testing and Makefile and other scaffolding
 * [X] Clean up and simplify tests
 * [X] Document resource in README.md
 * [X] Add example Dockerfile
 * [X] Read config from `/opt/resource/config.yml`
 * [X] Pass request down to the command
 * [X] Read from stdout the response
 * [X] Generic request and response
 * [X] Filter the request to the command
 * [ ] ... Auto filter params
 * [ ] ... test that filtered params are not passed
 * [ ] ... Test complex parameters with json
 * [ ] ... Test override of params
 * [ ] Add some examples to README.md
 * [ ] Optional redirect all output to stderr.
 * [ ] Print output to console, at least on error.
 * [ ] One line shell commands
 * [ ] Stdout/Stderr is captured and printed immediatelly (e.g. https://github.com/kvz/logstreamer)
 * [ ] Options how to capture stdout/stderr: "both-prefix|both|stdout|stderr"
 * [ ] Write raw requests and read raw responses as json from `SMUGGLER_INPUT_DIR` and `SMUGGLER_OUTPUT_DIR`

# Desired

 * [ ] Wrap other resources
 * [ ] Resource to "Inject smuggler" in other images, based in smuggler
 * [ ] smuggler for go inline code :)
 * [ ] autobuild docker
 * [ ] multiflavour docker (alpine, ubuntu, python, ruby, perl...)
 * [ ] add `source.default_check_version` to keep check version constant
 * [ ] Better error messages if config syntax is not right: Currently: `error reading request from stdin: json: cannot unmarshal object into Go value of type []smuggler.CommandDefinition
[0m`
 * [ ] Metadata file lines with json?

# Future ideas

 * [ ] Library to implement resources
 * [ ] Hooks for smuggler

# Smuggling ideas

  * [ ] GPG/SSH/Certificate generator
  * [ ] S3 init
  * [ ] Git tagger
  * [ ] bzr/subversion/any CVS...
  * [ ] good archive resource (examples combining it with S3)
  * [ ] FTP resource
  * [ ] Wrap other resources and do string interpolation in their parameters. For instance query a key/value store, with other resource to set variables :).

