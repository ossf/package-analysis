# Analysis runner.
import argparse
import json
import os
import subprocess
import urllib.parse
import urllib.request
from xml.etree import ElementTree

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
  url = f'https://pypi.org/rss/project/{safe_pkg}/releases.xml'
  resp = urllib.request.urlopen(url)
  root = ElementTree.fromstring(resp.read())
  return [next(e.itertext()) for e in root.findall('.//item/title')]


def _rubygems_versions_for_package(pkg):
  safe_pkg = urllib.parse.quote_plus(pkg)
  url = f'https://rubygems.org/api/v1/versions/{safe_pkg}.json'
  resp = urllib.request.urlopen(url)
  data = json.loads(resp.read())
  return [v['number'] for v in data]


def _versions_for_package(ecosystem, pkg):
    return {
        'npm': _npm_versions_for_package,
        'pypi': _pypi_versions_for_package,
        'rubygems': _rubygems_versions_for_package,
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

  if results_bucket:
    attributes.append('results_bucket_override=' + results_bucket)

  print('Requesting analysis with', attributes)
  topic = _TOPIC[_TOPIC.find('://') + 3:]
  subprocess.run(
    ('gcloud', 'pubsub', 'topics', 'publish',
      topic, '--attribute=' + ','.join(attributes)), check=True)


def main():
  parser = argparse.ArgumentParser(description='Analysis runner')
  parser.add_argument(
      'ecosystem', choices=('npm', 'pypi', 'rubygems'),
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
  parser.add_argument(
      '-b', '--results', help='Results bucket (overrides default).')

  args = parser.parse_args()

  if args.file and (not args.name or not args.version):
    raise ValueError('Need to specify package name and version for local file.')

  if args.list:
    with open(args.list) as f:
      for line in f.readlines():
        pkg, *rest = line.strip().split(' ')
        version = rest[0] if rest else args.version
        _request(
            pkg, args.ecosystem, version,
            results_bucket=args.results)

  elif args.name:
    if args.all:
      for version in _versions_for_package(args.ecosystem, args.name):
        _request(args.name, args.ecosystem, version,
            results_bucket=args.results)
    else:
      _request(
          args.name, args.ecosystem, args.version, local_file=args.file,
          results_bucket=args.results)


if __name__ == '__main__':
  main()
