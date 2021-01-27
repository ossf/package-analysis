#!/bin/bash

for pkg in $(cat node.txt); do
    kubectl run $pkg --image=node --restart='Never' --labels=install=1 --requests="cpu=250m" -- sh -c "mkdir -p /app && cd /app && npm init -f && npm install $pkg --save"
done
