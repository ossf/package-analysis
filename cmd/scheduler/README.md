# Scheduler

This directory contains code to schedule analysis jobs based on new package data from
our package feeds.

## Overview

This is a Golang app runs on Kubernetes and is deployed with [ko](https://github.com/google/ko).
This is currently running in a GKE cluster, if you need access reach out to dlorenc@google.com.

## Design

The goal is to create a set of Pods to be analyzed whenever there is a new package ready.
Each package can result in a specific set of analysys runs.

Right now, we only do a basic install for each package uploaded to the following package registries:
- [`PyPI`](https://pypi.org/)
- [`npmjs`](https://registry.npmjs.org/)
- [`RubyGems`](https://rubygems.org/)

We intend to support more language's package managers like [`cargo`](https://crates.io/) and other analysis types like
import-time monitoring and more!
