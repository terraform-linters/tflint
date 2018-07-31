# Package Managers

To create the scripts in `install-consul`, we had to find a way to write and package the scripts that satisfied a 
number of requirements. This document captures the requirements, the options we considered, and an explanation of 
which option we picked and why.



## The requirements

We need to write and package the scripts in this Module in a way that satisfies the following requirements:

- **Packages**. There needs to be a way to fetch these scripts from a canonical location (e.g. GitHub repo, package 
  manager repository) at a specific version number (e.g. `v0.0.3` of `install-consul`), much like a package manager. 
  We don't want people copy/pasting these scripts into their local repos, or it'll make upgrades and maintenance 
  difficult. 

- **Cross-platform**. The packaging system should work on most major Linux distributions. It should also work on OS X,
  as that's what many people use for development.

- **Handles dependencies**. These scripts rely on certain dependencies being installed on the system, such as `curl`,
  `wget`, `jq`, `aws`, and so on. We need a way to automatically manage and install these dependencies that works
  across all major Linux distributions. 

- **Simple package manager installation**: We don't want a package manager that takes a dozen steps to install. 

- **Simple client usage**. The scripts in this Module are fairly simple, so it shouldn't take a dozen steps to 
  install one. Ideally, we can use a one-liner such as `apt install -y install-consul`, except it should work on all 
  major Linux distributions.

- **Simple publish usage**. We need a fast and reliable way to publish new versions of the scripts. Ideally, we'd avoid
  having to publish each update to multiple package repos (apt, yum, etc), especially if that requires any sort of 
  manual approval (e.g. a PR for each new version). 

- **Testable in dev mode**. We must be able to do local, iterative development on the example code in the 
  [examples](https://github.com/hashicorp/terraform-aws-consul/tree/master/examples) folder. That means there is a way to "package" these scripts so that, in dev mode, they are
  downloaded from the local file system, rather than some package repo such as apt or yum.

- **Mature**: We want to use a solution that is mature, battle-tested, and has an active community around it. 



## The options

Here are the options we've looked at.

### [Nix](https://nixos.org/nix/)

- **Description**: Purely functional package manager, so dependency versioning, rollback, etc works very cleanly.
- **Dependency Management**: Yes.
- **Install process**: Simple. `bash <(curl https://nixos.org/nix/install)`.
- **Client usage**: Simple. `nix-env --install PACKAGE`.
- **Publish usage**: Complicated. Nix has its own [expression 
  language](https://nixos.org/nix/manual/#sec-expression-syntax), which I found fairly confusing. The docs are
  so-so. Creating new packages and pushing new versions seems to require [a pull 
  request](https://nixos.org/wiki/Create_and_debug_nix_packages).
- **Dev mode**: Complicated. Not clear how to use it in dev mode.  
- **Maturity**: Moderate. It's been around a while and there is a community around it, but it's buggy and confusing to 
  use.  

**Verdict**: It's confusing to use, slow (every install downloads the universe), buggy on OS X, and it's not clear how
to use it in dev mode. 

### [tpkg](http://tpkg.github.io/)
 
- **Description**: Package apps as super-powered tar files.
- **Dependency Management**. Yes. Supports both native installers (e.g. apt) and tpkg itself.
- **Install process**: Difficult. Requires Ruby and Ruby Gems to be installed first, so every Packer template would 
  have to install Ruby and Gem (e.g. `sudo apt install ruby`), which some people won't want on their production 
  servers, and then install `tpkg`: `sudo gem install tpkg`. 
- **Client usage**: Simple. `tpkg --install PACKAGE`.
- **Publish usage**: Simple. `tpkg --make PATH`. That produces a file you can upload to your own [package 
  server](http://tpkg.github.io/package_server.html), which can be any web server that hosts the file and a special
  metadata file. Might be able to use GitHub releases or S3 for this.
- **Dev mode**: Simple. The `tpkg --make PATH` command makes the package available for local install.   
- **Maturity**: Poor. Might be a dead project. The [GitHub repo](https://github.com/tpkg) has almost no followers. 
  Only a couple commits in the last few years.

**Verdict**: Dependency on Ruby and the lack of community activity is a no-go.

### [Snap](https://snapcraft.io/)

- **Description**: A way to install "apps" on all major Linux distributions. It seems like it's designed for standalone apps and
  binaries rather than scripting. Packages, called "snaps", are completely isolated from each other and the host OS 
  (using cgroups?) and can define interfaces, slots, plugs, etc to communicate with each other (a bit like "type 
  safety").
- **Dependency Management**: No. Or at least, I can't find it.
- **Install process**: Simple. `sudo apt install snapd`. 
- **Client usage**: Simple. `sudo snap install PACKAGE`.
- **Publish usage**: Complicated. You have to sign up for an account in the [Ubuntu 
  Store](https://myapps.developer.ubuntu.com/), install a separate app (`sudo apt install snapcraft`), login
  (`snapcraft login`), configure channels (stable, beta, etc); after that, it's an easy `snapcraft push` command 
  for each new version.
- **Dev mode**: Simple. `snapcraft` supports it.
- **Maturity**: Moderate. Community seems fairly active, as this is a project maintained by Canonical.  

**Verdict**: It only works on Linux, so hard to do development.

### `curl | bash`

- **Description**: Upload our scripts to Git, release them with version numbers, and pipe `curl` into `bash` to run them.
- **Dependency Management**: No.
- **Install process**: Simple. Nothing to install! Well, perhaps `curl`, but that's as simple as it gets.
- **Client usage**: Simple. `curl -Ls https://raw.githubusercontent.com/foo/bar/v0.0.3/install-consul | bash /dev/stdin`.
  Unfortunately, without any checksum or signature verification, this is a mild security risk if the GitHub repo 
  gets hijacked. Moreover, this only works for individual files. If the script has dependencies, those have to
  be downloaded separately.
- **Publish usage**: Simple. Just create a new GitHub release.
- **Dev mode**: Simple. Just change the URL to a local file path.
- **Maturity**: Strong. No need for a community, as we're just using `curl`!   

**Verdict**: This only works well for a single file. Of course, that file could download other files, but to do that,
the file has to know what version it is, what to use to download, where to download to, etc.


### [Gruntwork Installer](https://github.com/gruntwork-io/gruntwork-installer)

- **Description**: A slightly more structured version of piping `curl` into `bash`. You specify a GitHub repo, a path, and a version 
  number and the installer checks out the repo at the specified version, and runs an `install.sh` script in the
  specified path.
- **Dependency Management**: No. It's up to the `install.sh` script to figure out the details.
- **Install process**: Simple. `curl -Ls https://raw.githubusercontent.com/gruntwork-io/gruntwork-installer/master/bootstrap-gruntwork-installer.sh | bash /dev/stdin --version v0.0.14`.
  Note, this is subject to the same security risks as piping `curl` into `bash`. Since there is just one installer
  and we don't update it often, we *could* publish it into apt, yum, etc repos to avoid this problem.
- **Client usage**: Simple. `gruntwork-install --module-name 'PATH' --repo 'https://github.com/foo/bar' --tag v0.0.3`.
  Does not currently do checksum or signature verification, but that could be added.
- **Publish usage**: Simple. Just create a new GitHub release. Works with private GitHub repos too.
- **Dev mode**: Simple. Just specify a local file path.
- **Maturity**: Poor. The community is tiny, though this project is actively maintained by Gruntwork.

**Verdict**: A bit too specific to Gruntwork's use case.

### [fpm](https://github.com/jordansissel/fpm)

- **Description**: A script that makes it easy to package your code as native packages (e.g. `.deb`, `.rpm`).
- **Dependency Management**: Yes. 
- **Install process**: Simple. No install process, as you use your standard OS package managers (i.e. `apt`, `yum`).
- **Client usage**: Simple. `sudo apt install -y PACKAGE`.
- **Publish usage**: Complicated. You have to package and publish to all major Linux package repos.
- **Dev mode**: Complicated. Not clear how you use it in dev mode.
- **Maturity**: Strong. Big community, active project.

**Verdict**: Requires publishing to multiple repos for every release, which is complicated. 

### Configuration management tools (e.g. [Ansible](https://www.ansible.com/), [Chef](https://www.chef.io/))

- **Description**: Tools built for managing server configuration.
- **Dependency Management**: Yes. Most cfg mgmt tools have ways of leveraging the built-in package managers
  (e.g. [package command in Chef](https://docs.chef.io/resource_package.html) and [package command in 
  Ansible](http://docs.ansible.com/ansible/package_module.html)).
- **Install process**: Simple. Packer can do it automatically for [Chef 
  Solo](https://www.packer.io/docs/provisioners/chef-solo.html) and you can do it manually for Ansible:
  `sudo apt install -y ansible`.
- **Client usage**: Complicated. You first have to download the Chef Recipe or Ansible Playbook from the Module
  repo (e.g. using a `shell-local` provisioner with `curl`) and then you can use the downloaded recipe or playbook 
  with the built-in Packer commands (e.g. [chef-solo 
  Provisioner](https://www.packer.io/docs/provisioners/chef-solo.html) and [ansible-local 
  Provisioner](https://www.packer.io/docs/provisioners/ansible-local.html)).
- **Publish usage**: Simple. Just create a new GitHub release.
- **Dev mode**: Simple. Just use local file paths for the recipes and playbooks.
- **Maturity**: Strong. All these cfg mgmt tools have massive communities.

**Verdict**: Requires installing the tools on each server and learning a new set of tools, which feels like overkill
for a few simple scripts.

### Git

- **Description**: Run `git clone` with the `--branch` parameter (which can be set to a tag) to check out a specific version of the 
  code.
- **Dependency Management**: No.
- **Install process**: Simple. Just install Git, if it's not installed already.
- **Client usage**: Simple. Once you've run `git clone`, all the code you need is on disk, and you just execute it.
- **Publish usage**: Simple. Just create a new GitHub release.
- **Dev mode**: Simple. Just use your local checkout.
- **Maturity**: Strong. It's Git, so the community is massive.

**Verdict**: The biggest missing feature is dependency management, but it's a perfect fit in every other way, so 
this is our choice.