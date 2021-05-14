#!/bin/bash
topic=${OSSMALWARE_WORKER_TOPIC:-projects/ossf-malware-analysis/topics/workers}
for pkg in $(cat python.txt); do
  gcloud pubsub topics publish "$topic" --attribute=name=$pkg,ecosystem=pypi
done
