# Analysis runner.
import argparse
import os
import subprocess

_TOPIC = os.getenv(
    'OSSMALWARE_WORKER_TOPIC',
    'gcppubsub://projects/ossf-malware-analysis/topics/workers')
_PACKAGES_BUCKET = os.getenv(
    'OSSF_MALWARE_ANALYSIS_PACKAGES', 'gs://ossf-malware-analysis-packages')


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
  parser.add_argument('-b', '--results', help='Results bucket (overrides default).')

  args = parser.parse_args()

  if args.file and (not args.name or not args.version):
    raise ValueError('Need to specify package name and version for local file.')

  if args.list:
    with open(args.list) as f:
      for line in f.readlines():
        _request(
            line.strip(), args.ecosystem, args.version,
            results_bucket=args.results)

  elif args.name:
    _request(
        args.name, args.ecosystem, args.version, local_file=args.file,
        results_bucket=args.results)


if __name__ == '__main__':
  main()
