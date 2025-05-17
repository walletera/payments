test_privateapi:
	go test --tags=privateapi -count=1 -v ./...

wait_containers_removed:
	while docker ps -a --quiet | grep -q .; do sleep 1; done

test_publicapi:
	go test --tags=publicapi -count=1 -v ./...

test_all: | test_privateapi wait_containers_removed test_publicapi