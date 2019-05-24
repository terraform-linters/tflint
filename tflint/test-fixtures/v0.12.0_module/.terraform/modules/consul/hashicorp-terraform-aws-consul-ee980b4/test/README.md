# Tests

This folder contains automated tests for this Module. All of the tests are written in [Go](https://golang.org/). 
Most of these are "integration tests" that deploy real infrastructure using Terraform and verify that infrastructure 
works as expected using a helper library called [Terratest](https://github.com/gruntwork-io/terratest).  



## WARNING WARNING WARNING

**Note #1**: Many of these tests create real resources in an AWS account and then try to clean those resources up at 
the end of a test run. That means these tests may cost you money to run! When adding tests, please be considerate of 
the resources you create and take extra care to clean everything up when you're done!

**Note #2**: Never forcefully shut the tests down (e.g. by hitting `CTRL + C`) or the cleanup tasks won't run!

**Note #3**: We set `-timeout 60m` on all tests not because they necessarily take that long, but because Go has a
default test timeout of 10 minutes, after which it forcefully kills the tests with a `SIGQUIT`, preventing the cleanup
tasks from running. Therefore, we set an overlying long timeout to make sure all tests have enough time to finish and 
clean up.



## Running the tests

### Prerequisites

- Install the latest version of [Go](https://golang.org/).
- Install [glide](https://glide.sh/) for Go dependency management.
- Install [Terraform](https://www.terraform.io/downloads.html).
- Configure your AWS credentials using one of the [options supported by the AWS 
  SDK](http://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html). Usually, the easiest option is to
  set the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.


### One-time setup

Download Go dependencies using glide:

```
cd test
glide install
```


### Run all the tests

```bash
cd test
go test -v -timeout 60m
```


### Run a specific test

To run a specific test called `TestFoo`:

```bash
cd test
go test -v -timeout 60m -run TestFoo
```


  