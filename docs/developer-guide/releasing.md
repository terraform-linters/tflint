# Releasing a new version of TFLint

Maintainers with push access to TFLint can publish new releases.
Keep the following in mind when publishing a release:

- Avoid changing behavior in patch versions as much as possible. Patch version updates should only contain bug fixes.
  - However, if the behavior changes due to bug fixes, this is not the case.
- Write readable release notes. It would be nice to know what the benefits are of updating to this version, or what needs to be changed to update.

To do this efficiently, releasing is automated as much as possible. After merging the changes you want to release into the master branch, run `make release` on the master branch.

```console
$ make release
```

The `make release` does the following:

- Accept new version as input and automatically rewrites the necessary files.
- Automatically generate a release note. You can edit it interactively.
- Run the required tests.
- Create a commit, tag it and push it to the remote.

Note that this script does not build any binaries. These run on GitHub Actions. This script just pulls the trigger for it.
