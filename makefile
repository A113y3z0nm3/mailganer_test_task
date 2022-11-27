.PHONY: docker-build docker-run

docker-build:
	docker build -t mailganer_test_task  .

docker-run:
	docker run --network host -it --rm \
	--env-file ./configs/.env mailganer_test_task:latest