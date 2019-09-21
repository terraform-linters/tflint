# Credentials

In [Deep checking](advanced.md#deep-checking), it is necessary to set provider's credentials in order to call APIs. Currently, only AWS is supported.

Credentials are used with the following priority:

- Static credentials
- Static credentials (Terraform)
- Environment variables
- Shared credentials
- Shared credentials (Terraform)
- ECS and CodeBuild task roles
- EC2 role

## Static credentials

If you have an access key and a secret key, you can pass these keys like the following:

```
$ tflint --aws-access-key AWS_ACCESS_KEY --aws-secret-key AWS_SECRET_KEY --aws-region us-east-1
```

```hcl
config {
  aws_credentials = {
    access_key = "AWS_ACCESS_KEY"
    secret_key = "AWS_SECRET_KEY"
    region     = "us-east-1"
  }
}
```

Although there is not recommended, if an access key is hard-coded in a provider definition, they will also be taken into account. However, aliases are not supported. The priority is higher than the environment variable and lower than the above way.

```hcl
provider "aws" {
  region     = "us-west-2"
  access_key = "my-access-key"
  secret_key = "my-secret-key"
}
```

## Shared credentials

If you have [shared credentials](https://aws.amazon.com/jp/blogs/security/a-new-and-standardized-way-to-manage-credentials-in-the-aws-sdks/), you can pass a profile name and credentials file path. If omitted, these will be `default` and `~/.aws/credentials`.

```
$ tflint --aws-profile AWS_PROFILE --aws-region us-east-1 --aws-creds-file ~/.aws/myapp
```

```hcl
config {
  aws_credentials = {
    profile                 = "AWS_PROFILE"
    region                  = "us-east-1"
    shared_credentials_file = "~/.aws/myapp"
  }
}
```

If these configurations are defined in the provider block, they will also be taken into account. But the priority is lower than the above way.

```hcl
provider "aws" {
  region                  = "us-west-2"
  shared_credentials_file = "/Users/tf_user/.aws/creds"
  profile                 = "customprofile"
}
```

## Environment variables

TFLint looks up `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` environment variables. This is useful when you don't want to explicitly pass credentials.

```
$ export AWS_ACCESS_KEY_ID=AWS_ACCESS_KEY
$ export AWS_SECRET_ACCESS_KEY=AWS_SECRET_KEY
```

## Role-based authentication

TFLint fetches AWS credentials in the same way as Terraform. See [this documentation](https://www.terraform.io/docs/providers/aws/index.html#ecs-and-codebuild-task-roles) for role-based authentication.

## Assume role

TFLint can assume a role in the same way as Terraform. See [this documentation](https://www.terraform.io/docs/providers/aws/index.html#assume-role).
