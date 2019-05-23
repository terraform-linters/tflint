# Contribution Guidelines

Contributions to this Module are very welcome! We follow a fairly standard [pull request 
process](https://help.github.com/articles/about-pull-requests/) for contributions, subject to the following guidelines:
 
1. [File a GitHub issue](#file-a-github-issue)
1. [Update the documentation](#update-the-documentation)
1. [Update the tests](#update-the-tests)
1. [Update the code](#update-the-code)
1. [Create a pull request](#create-a-pull-request)
1. [Merge and release](#merge-and-release)

## File a GitHub issue

Before starting any work, we recommend filing a GitHub issue in this repo. This is your chance to ask questions and
get feedback from the maintainers and the community before you sink a lot of time into writing (possibly the wrong) 
code. If there is anything you're unsure about, just ask!

## Update the documentation

We recommend updating the documentation *before* updating any code (see [Readme Driven 
Development](http://tom.preston-werner.com/2010/08/23/readme-driven-development.html)). This ensures the documentation 
stays up to date and allows you to think through the problem at a high level before you get lost in the weeds of 
coding.

## Update the tests

We also recommend updating the automated tests *before* updating any code (see [Test Driven 
Development](https://en.wikipedia.org/wiki/Test-driven_development)). That means you add or update a test case, 
verify that it's failing with a clear error message, and *then* make the code changes to get that test to pass. This 
ensures the tests stay up to date and verify all the functionality in this Module, including whatever new 
functionality you're adding in your contribution. Check out the [tests](https://github.com/hashicorp/terraform-aws-consul/tree/master/test) folder for instructions on running the 
automated tests. 

## Update the code

At this point, make your code changes and use your new test case to verify that everything is working. As you work,
keep in mind two things:

1. Backwards compatibility
1. Downtime

### Backwards compatibility

Please make every effort to avoid unnecessary backwards incompatible changes. With Terraform code, this means:

1. Do not delete, rename, or change the type of input variables.
1. If you add an input variable, it should have a `default`.
1. Do not delete, rename, or change the type of output variables.
1. Do not delete or rename a module in the `modules` folder.

If a backwards incompatible change cannot be avoided, please make sure to call that out when you submit a pull request, 
explaining why the change is absolutely necessary. 

### Downtime

Bear in mind that the Terraform code in this Module is used by real companies to run real infrastructure in 
production, and certain types of changes could cause downtime. For example, consider the following:

1. If you rename a resource (e.g. `aws_instance "foo"` -> `aws_instance "bar"`), Terraform will see that as deleting
   the old resource and creating a new one.
1. If you change certain attributes of a resource (e.g. the `name` of an `aws_elb`), the cloud provider (e.g. AWS) may
   treat that as an instruction to delete the old resource and a create a new one. 
   
Deleting certain types of resources (e.g. virtual servers, load balancers) can cause downtime, so when making code
changes, think carefully about how to avoid that. For example, can you avoid downtime by using 
[create_before_destroy](https://www.terraform.io/docs/configuration/resources.html#create_before_destroy)? Or via
the `terraform state` command? If so, make sure to note this in our pull request. If  downtime cannot be avoided, 
please make sure to call that out when you submit a pull request. 

## Create a pull request

[Create a pull request](https://help.github.com/articles/creating-a-pull-request/) with your changes. Please make sure
to include the following:

1. A description of the change, including a link to your GitHub issue.
1. The output of your automated test run, preferably in a [GitHub Gist](https://gist.github.com/). We cannot run 
   automated tests for pull requests automatically due to [security 
   concerns](https://circleci.com/docs/fork-pr-builds/#security-implications), so we need you to manually provide this 
   test output so we can verify that everything is working.
1. Any notes on backwards incompatibility or downtime.

## Merge and release   

The maintainers for this repo will review your code and provide feedback. If everything looks good, they will merge the
code and release a new version, which you'll be able to find in the [releases page](../../releases).