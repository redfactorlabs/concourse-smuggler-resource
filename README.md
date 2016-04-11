# concourse-smuggler-resource

Concourse generic resource, which would allow quickly implement any resource

This is currently a WIP idea with no functional code.

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

## Example

To add this resource using `custom_resource`

```
resource_types:
- name: smuggler
  type: docker-image
  source:
    repository: redfactorlabs/concourse-smuggler-resource
```

Defining a resource that creates a `id_rsa` key and stores it in S3

```
resources:
- name: atomy-pr
  type: pull-request
  source:
    backend:
    - name: s3
      config:
        bucket_name: my-ssh-key-store
        versioned_file: id_rsa.tar.gz
    commands:
    - name: out
      path: sh
      args:
      - -e
      - -c
      - |
        ssh-keygen -f id_rsa -N ''
        tar -cvzf ./backend/s3/id_rsa.tar.gz id_rsa id_rsa.tgz
```

# TODO

 * [ ] basic check/in/out which runs a script command

