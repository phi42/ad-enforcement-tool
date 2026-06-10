# ADE CI templates

Ready-to-use CI workflow templates that run `ade` against a project on every push and pull request. Copy the template for your CI provider into your own repository under the corresponding workflows directory and adapt the placeholders to your project layout.

## Available templates

| Directory                            | Provider       | Description                                                                                          |
| ------------------------------------ | -------------- | ---------------------------------------------------------------------------------------------------- |
| [`github-actions/`](github-actions/) | GitHub Actions | Installs `ade` and one or more plugins, then runs `verify` (and optionally `compile`) on every push. |

See the [user guide](../../docs/user-guide.md#step-6-run-ade-in-ci) for the surrounding workflow, including the considerations behind the example below.
