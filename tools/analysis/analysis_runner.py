# Analysis runner.
import argparse
import json
import os
import subprocess
import urllib.parse
import urllib.request

_ECOSYSTEMS = ('npm', 'pypi', 'rubygems', 'packagist', 'crates.io')
_TOPIC = os.getenv(
    'OSSMALWARE_WORKER_TOPIC',
    'gcppubsub://projects/ossf-malware-analysis/topics/workers')
_PACKAGES_BUCKET = os.getenv(
    'OSSF_MALWARE_ANALYSIS_PACKAGES', 'gs://ossf-malware-analysis-packages')
_NPM_IGNORE_KEYS = ('modified', 'created')


def _npm_versions_for_package(pkg):
  safe_pkg = urllib.parse.quote_plus(pkg)
  url = f'https://registry.npmjs.com/{safe_pkg}'
  resp = urllib.request.urlopen(url)
  data = json.loads(resp.read())
  versions = data.get('time', {}).keys()
  return [v for v in versions if v not in _NPM_IGNORE_KEYS][::-1]


def _pypi_versions_for_package(pkg):
  safe_pkg = urllib.parse.quote_plus(pkg)
  url = f'https://pypi.org/pypi/{safe_pkg}/json'
  resp = urllib.request.urlopen(url)
  data = json.loads(resp.read())
  releases = data.get('releases', {})
  return [v for v, d in releases.items() if d][::-1]


def _rubygems_versions_for_package(pkg):
  safe_pkg = urllib.parse.quote_plus(pkg)
  url = f'https://rubygems.org/api/v1/versions/{safe_pkg}.json'
  resp = urllib.request.urlopen(url)
  data = json.loads(resp.read())
  return [v['number'] for v in data]


def _packagist_versions_for_package(pkg):
  safe_pkg = urllib.parse.quote_plus(pkg)
  url = f'https://repo.packagist.org/p2/{safe_pkg}.json'
  resp = urllib.request.urlopen(url)
  data = json.loads(resp.read())
  packages = data.get('packages', {})
  return [v['version'] for p in packages.values() for v in p]


def _crates_versions_for_package(pkg):
  safe_pkg = urllib.parse.quote_plus(pkg)
  url = f'https://crates.io/api/v1/crates/{safe_pkg}/versions'
  resp = urllib.request.urlopen(url)
  data = json.loads(resp.read())
  versions = data.get('versions', [])
  return [v['num'] for v in versions]


def _versions_for_package(ecosystem, pkg):
    return {
        'npm': _npm_versions_for_package,
        'pypi': _pypi_versions_for_package,
        'rubygems': _rubygems_versions_for_package,
        'packagist': _packagist_versions_for_package,
        'crates.io': _crates_versions_for_package,
    }[ecosystem](pkg)


def _upload_file(local_path):
  # TODO: figure out a better way to key these.
  filename = os.path.basename(local_path)
  upload_path = f'{_PACKAGES_BUCKET}/{filename}'
  print('Uploading', local_path, 'to', upload_path)
  subprocess.run(
    ('gsutil', 'cp', local_path, upload_path), check=True)

  return filename


def _request(name, ecosystem, version, local_file=None, results_bucket=None):
  attributes = [
    'name=' + name,
    'ecosystem=' + ecosystem,
  ]

  if version:
    attributes.append('version=' + version)

  if local_file:
    uploaded_path = _upload_file(local_file)
    attributes.append('package_path=' + uploaded_path)

  print('Requesting analysis with', attributes)
  topic = _TOPIC[_TOPIC.find('://') + 3:]
  subprocess.run(
    ('gcloud', 'pubsub', 'topics', 'publish',
      topic, '--attribute=' + ','.join(attributes)), check=True)


def main():
  parser = argparse.ArgumentParser(description='Analysis runner')
  parser.add_argument(
      'ecosystem', choices=_ECOSYSTEMS,
      help='Package ecosystem')

  group = parser.add_mutually_exclusive_group(required=True)
  group.add_argument('-l', '--list', help='List of package names as a file.')
  group.add_argument('-n', '--name', help='Package name')

  parser.add_argument('-f', '--file', help='Local package file')
  parser.add_argument('-v', '--version', help='Package version')
  parser.add_argument(
      '-a', '--all', default=False,
      action=argparse.BooleanOptionalAction,
      help='Use all publised versions for the package')

  args = parser.parse_args()

  local_file = None
  if args.file:
    if not args.name or not args.version:
      raise ValueError(
          'Need to specify package name and version for local file.')

    local_file = args.file

  package_names = []
  if args.list:
    with open(args.list) as f:
      package_names = [line.strip() for line in f.readlines() if line.strip()]
  elif args.name:
    package_names = [args.name]

  if not package_names:
    raise ValueError('No package name found.')

  for package in package_names:
    if args.all:
      for version in _versions_for_package(args.ecosystem, package):
        _request(
            package, args.ecosystem, version,
            results_bucket=args.results)
    else:
      _request(
          package, args.ecosystem, args.version,
          local_file=local_file,
          results_bucket=args.results)


if __name__ == '__main__':
  main()
