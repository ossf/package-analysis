#!/bin/bash
topic=${OSSMALWARE_WORKER_TOPIC:-projects/ossf-malware-analysis/topics/workers}
for pkg in $(cat node.txt); do
  gcloud pubsub topics publish "$topic" --attribute=name=$pkg,ecosystem=npm
done
