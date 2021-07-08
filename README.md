## Github PRs resource

[graphql-api]: https://developer.github.com/v4
[original-resource]: https://github.com/telia-oss/github-pr-resource
[git-resource]: https://github.com/concourse/git-resource
[#example]: #example

A proof-of-concept Concourse resource type for tracking the list of pull requests to a GitHub repository.
The code is adapted from [the original][original-resource], with some important distinctions:

* This resource emits a single version consisting of a list of PR numbers,
  whereas [the original] emits a version for each new commit in each PR.
    * This list of PR numbers can be paired with the `load_var` step, the
      `across` step, and the `set_pipeline` step to dynamically construct a
      group of instanced pipelines (each PR gets its own ephemeral instanced
      pipeline)
* This resource does not handle tracking/fetching commits to PRs
    * The intention is that you'll use this resource type in conjunction with
      the built-in [git-resource]

Refer to [#example] for a full example.

## Source Configuration

| Parameter                   | Required | Example                          | Description                                                                                                                                                                                                                                                                                 |
|-----------------------------|----------|----------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `repository`                | Yes      | `itsdalmo/test-repository`       | The repository to target.                                                                                                                                                                                                                                                                   |
| `access_token`              | Yes      |                                  | A Github Access Token with repository access (required for setting status on commits). N.B. If you want github-prs-resource to work with a private repository. Set `repo:full` permissions on the access token you create on GitHub. If it is a public repository, `repo:status` is enough. |
| `v3_endpoint`               | No       | `https://api.github.com`         | Endpoint to use for the V3 Github API (Restful).                                                                                                                                                                                                                                            |
| `v4_endpoint`               | No       | `https://api.github.com/graphql` | Endpoint to use for the V4 Github API (Graphql).                                                                                                                                                                                                                                            |
| `paths`                     | No       | `["terraform/*/*.tf"]`           | Only produce new versions if the PR includes changes to files that match one or more glob patterns or prefixes.                                                                                                                                                                             |
| `ignore_paths`              | No       | `[".ci/"]`                       | Inverse of the above. Pattern syntax is documented in [filepath.Match](https://golang.org/pkg/path/filepath/#Match), or a path prefix can be specified (e.g. `.ci/` will match everything in the `.ci` directory).                                                                          |
| `disable_ci_skip`           | No       | `true`                           | Disable ability to skip builds with `[ci skip]` and `[skip ci]` in commit message or pull request title.                                                                                                                                                                                    |
| `skip_ssl_verification`     | No       | `true`                           | Disable SSL/TLS certificate validation on API clients. Use with care!                                                                                                                                                                                                                       |
| `disable_forks`             | No       | `true`                           | Disable triggering of the resource if the pull request's fork repository is different to the configured repository.                                                                                                                                                                         |
| `ignore_drafts`             | No       | `false`                          | Disable triggering of the resource if the pull request is in Draft status.                                                                                                                                                                                                                  |
| `required_review_approvals` | No       | `2`                              | Disable triggering of the resource if the pull request does not have at least `X` approved review(s).                                                                                                                                                                                       |
| `base_branch`               | No       | `master`                         | Name of a branch. The pipeline will only trigger on pull requests against the specified branch.                                                                                                                                                                                             |
| `labels`                    | No       | `["bug", "enhancement"]`         | The labels on the PR. The pipeline will only trigger on pull requests having at least one of the specified labels.                                                                                                                                                                          |
| `disable_git_lfs`           | No       | `true`                           | Disable Git LFS, skipping an attempt to convert pointers of files tracked into their corresponding objects when checked out into a working copy.                                                                                                                                            |
| `states`                    | No       | `["OPEN", "MERGED"]`             | The PR states to select (`OPEN`, `MERGED` or `CLOSED`). The pipeline will only trigger on pull requests matching one of the specified states. Default is ["OPEN"].                                                                                                                          |
| `number`                    | No       | `1234`                           | The PR number to use by default for `put`. Has no effect for in `check` or `get`. The `put` parameter `params.number` takes precedence over `source.number`                                                                                                                                 |

Notes:
 - If `v3_endpoint` is set, `v4_endpoint` must also be set (and the other way around).
 - When using `required_review_approvals`, you may also want to enable GitHub's branch protection rules to [dismiss stale pull request approvals when new commits are pushed](https://help.github.com/en/articles/enabling-required-reviews-for-pull-requests).

## Behaviour

#### `check`

Produces a version consisting of the list of all PRs that match the criteria defined in the `source`, sorted by PR number.
A new version will only be emitted if the set of PRs has changed (i.e. PRs were added/removed from the set).

#### `get`

Stores the set of PR numbers in the file `prs.json`, encoded as a JSON list.
This file can then be loaded into the build's local var state via the `load_var` step.

Refer to [#example] for a full example.


#### `put`

| Parameter                  | Required | Example                              | Description                                                                                                                                                   |
|----------------------------|----------|--------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `path`                     | Yes      | `pull-request`                       | The name given to the resource in a GET step.                                                                                                                 |
| `number`                   | No       | `1234`                               | The PR number to modify. Must configure one of `params.number` or `source.number`. `params.number` takes precedence.                                          |
| `status`                   | No       | `SUCCESS`                            | Set a status on a commit. One of `SUCCESS`, `PENDING`, `FAILURE` and `ERROR`.                                                                                 |
| `base_context`             | No       | `concourse-ci`                       | Base context (prefix) used for the status context. Defaults to `concourse-ci`.                                                                                |
| `context`                  | No       | `unit-test`                          | A context to use for the status, which is prefixed by `base_context`. Defaults to `status`.                                                                   |
| `comment`                  | No       | `hello world!`                       | A comment to add to the pull request.                                                                                                                         |
| `target_url`               | No       | `$ATC_EXTERNAL_URL/builds/$BUILD_ID` | The target URL for the status, where users are sent when clicking details (defaults to the Concourse build page).                                             |
| `description`              | No       | `Concourse CI build failed`          | The description status on the specified pull request.                                                                                                         |
| `delete_previous_comments` | No       | `true`                               | Boolean. Previous comments made on the pull request by this resource will be deleted before making the new comment. Useful for removing outdated information. |

Note that `comment`, `context,` and `target_url` will all expand environment variables, so in the examples above `$ATC_EXTERNAL_URL` will be replaced by the public URL of the Concourse ATCs.
See https://concourse-ci.org/implementing-resource-types.html#resource-metadata for more details about metadata that is available via environment variables.

## Example

Unlike the [original resource][original-resource], usage of `aoldershaw/github-prs-resource`
requires two pipeline templates.

1. There is one "parent" pipeline that track changes to the list of PRs and
   sets a group of "child" pipelines. It runs whenever the list of PRs changes
   (e.g. because a new PR is opened, or an existing PR is merged).
2. There is one "child" pipeline per active PR. Each "child" pipeline tracks
   new commits to the corresponding PR. These pipelines will be archived when
   they are removed from the list of PRs (e.g. when the PR is merged).

For this example, assume you have a resource named `ci`, a repo which contains
the following pipeline files:

`ci/pipelines/parent.yml`
```yaml
resource_types:
- name: pull-requests
  type: registry-image
  source:
    repository: aoldershaw/github-prs-resource

resources:
- name: pull-requests
  type: pull-requests
  source:
    repository: itsdalmo/test-repository
    access_token: ((github-access-token))

- name: ci
  type: git
  source:
    uri: https://github.com/concourse/ci

jobs:
- name: update-pr-pipelines
  plan:
  - get: ci
  - get: pull-requests
    trigger: true
  - load_var: pull_requests
    file: pull-requests/prs.json
  - across:
    - var: pr_number
      values: ((.:pull_requests))
    set_pipeline: prs
    file: ci/pipelines/child.yml
    instance_vars: {number: ((.:pr_number))}
```

`ci/pipelines/child.yml`
```yaml
resource_types:
- name: pull-requests
  type: registry-image
  source:
    repository: aoldershaw/github-prs-resource

resources:
- name: pr
  type: git
  source:
    uri: https://((github-access-token))@github.com/itsdalmo/test-repository
    ref: pull/((number))/head

- name: pr-status
  type: pull-requests
  source:
    repository: itsdalmo/test-repository
    access_token: ((github-access-token))
    number: ((number))

- name: test
  plan:
  - get: pr
    trigger: true
  - put: pr-status
    params:
      path: pr
      status: pending
  - task: unit-test
    config:
      platform: linux
      image_resource:
        type: registry-image
        source: {repository: alpine/git, tag: "latest"}
      inputs:
        - name: pr
      run:
        path: /bin/sh
        args:
          - -xce
          - |
            cd pull-request
            git log --graph --all --color --pretty=format:"%x1b[31m%h%x09%x1b[32m%d%x1b[0m%x20%s" > log.txt
            cat log.txt
  - put: pr-status
    params:
      path: pr
      status: success
  on_failure:
    put: pr-status
    params:
      path: pr
      status: failure
```

## Costs

The Github API(s) have a rate limit of 5000 requests per hour (per user). For the V3 API this essentially
translates to 5000 requests, whereas for the V4 API (GraphQL)  the calculation is more involved:
https://developer.github.com/v4/guides/resource-limitations/#calculating-a-rate-limit-score-before-running-the-call

Ref the above, here are some examples of running `check` against large repositories and the cost of doing so:
- [concourse/concourse](https://github.com/concourse/concourse): 51 open pull requests at the time of testing. Cost 2.
- [torvalds/linux](https://github.com/torvalds/linux): 305 open pull requests. Cost 8.
- [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes): 1072 open pull requests. Cost: 22.

For the other two operations the costing is a bit easier:
- `get`: 0 cost
- `put`: Uses the V3 API and has a min cost of 1, +1 for each of `status`, `comment` etc.
