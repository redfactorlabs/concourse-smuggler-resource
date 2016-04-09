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


