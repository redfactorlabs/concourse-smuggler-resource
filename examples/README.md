Examples
========

Set of examples using smuggler.

Pre configuration of the environment
-----------------------------------

Providing your AWS credentials loaded in environment, you can run:

```
export AWS_DEFAULT_REGION=us-east-1

# Some "unique" part in the name
SUFFIX=$(echo $AWS_ACCESS_KEY_ID | shasum | cut -c1-8)

terraform apply -var suffix=${SUFFIX} shared/terraform

BUCKET_NAME=$(cat shared/terraform/terraform.tfstate | grep '"bucket":' | head -n1 | cut -f 4 -d \")
```

And go to any directory and run: `./build.sh`
