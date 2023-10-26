#! /bin/bash

if [[ -z ${GIT_TAG} ]]; then
	echo "Missing git tag"
	exit 1
fi

git checkout "${GIT_TAG}"
gcloud container clusters get-credentials analysis-cluster --zone=us-central1-c --project=ossf-malware-analysis

pushd infra/worker || echo "could not change to infra/worker directory" && exit 1

echo "Were any changes made to the k8s config?"
echo "Enter y to apply config changes and then restart workers, n to just restart, ctrl-C to exit"
read yn
case $yn in
	[Yy]* )
		echo "kubectl apply -f ."
		kubectl apply -f .
		;;
	[Nn]* )
		echo "kubectl rollout restart statefulset workers-deployment"
		kubectl rollout restart statefulset workers-deployment
	;;
esac

