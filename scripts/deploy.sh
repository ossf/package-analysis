#! /bin/bash

if [[ -z ${GIT_TAG} ]]; then
	echo "Missing git tag"
	exit 1
fi

echo "git checkout ${GIT_TAG}"
git checkout "${GIT_TAG}"

if ! git diff-index --quiet HEAD; then
	echo "there are uncommitted changes, please ensure the repo is clean"
	exit 1
fi

gcloud container clusters get-credentials analysis-cluster --zone=us-central1-c --project=ossf-malware-analysis

pushd infra/worker || (echo "pushd infra/worker failed" && exit 1)

echo "Were any changes made to the k8s config?"
echo "Enter y to apply config changes and then restart workers, n to just restart, ctrl-C to exit"
read -r yn
case $yn in
	[Yy]* )
		echo "kubectl apply -f $(pwd)"
		kubectl apply -f .
		;;
	[Nn]* )
		echo "kubectl rollout restart deployment workers-deployment"
		kubectl rollout restart statefulset workers-deployment
	;;
esac


popd || (echo "failed to popd" && exit 1)
