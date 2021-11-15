# Analysis Tools

## Analysis Runner

The `analysis-runner.sh` script is used to inject package names into the PubSub
queue the analysis pipeline consumes work from.

`node.txt`, `python.txt` and `rubygems.txt` contain a lists of the top packages
from these package repositories (at the time of creation). The data is from
[NPM](https://www.npmjs.com/browse/depended) (* dead),
[PyPI](https://hugovk.github.io/top-pypi-packages/top-pypi-packages-30-days.json)
and [RubyGems](https://rubygems.org/stats).

### Example usage

Firstly, ensure you are authenticated with the cluster:

```shell
$ gcloud container clusters get-credentials analysis-cluster \
    --zone=us-central1-c --project=ossf-malware-analysis
```

Here are some possible ways to invoke the script:

```shell
$ ./analysis-runner.sh pypi python.txt
$ ./analysis-runner.sh npm node.txt
$ ./analysis-runner.sh rubygems rubygems.txt
$ echo "my-npm-package" | ./analysis-runner.sh npm
```

