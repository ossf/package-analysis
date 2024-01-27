.PHONY: build_sample_python_package

IMAGE_NAME = sample-python-package-image
CONTAINER_NAME = sample-python-package-container

build_sample_python_package:
	docker build -t ${IMAGE_NAME} sample_python_package/
	docker run --name ${CONTAINER_NAME} -d ${IMAGE_NAME}
	docker cp ${CONTAINER_NAME}:/sample_python_package/dist/. sample_python_package/output
	docker stop ${CONTAINER_NAME}
	docker rm ${CONTAINER_NAME}
	docker image rm ${IMAGE_NAME}
