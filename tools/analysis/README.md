# Analysis Tools

## Analysis Runner

The `analysis_runner.py` script is used to inject packages into the PubSub
queue the analysis pipeline consumes work from.

`node.txt`, `python.txt` and `rubygems.txt` contain a lists of the top packages
from these package repositories (at the time of creation). The data is from
[NPM](https://www.npmjs.com/browse/depended) (* dead),
[PyPI](https://hugovk.github.io/top-pypi-packages/top-pypi-packages-30-days.json)
and [RubyGems](https://rubygems.org/stats).

### Prerequisites

This script requires:

- Python 3
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)

### Example usage

Firstly, ensure you are authenticated with the cloud project:

```shell
$ gcloud auth login
```

Here are some possible ways to invoke the script:

```shell
$ python3 analysis_runner.py pypi --list python.txt
$ python3 analysis_runner.py npm --list node.txt
$ python3 analysis_runner.py npm --name my-npm-package
$ python3 analysis_runner.py npm --name my-npm-package --version 0.1.1 --file /path/to/local.tgz
```

### Bulk backfill

To request a bulk backfill of a list of packages in a particular ecosystem:

```shell
$ ./backfill.sh <path/to/packages/delimited/by/newlines> <ecosystem>
```
