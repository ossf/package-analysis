#!/bin/bash

for pkg in $(cat node.txt); do
  rand=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c 10)
  kubectl run npm-$rand --image=node --restart='Never' \
      --labels=install=1,package_type=npm \
      --annotations=package_name=$pkg \
      --annotations=package_version=test \
      --requests="cpu=250m" -- sh -c "mkdir -p /app && cd /app && npm init -f && npm install $pkg --save"
done
