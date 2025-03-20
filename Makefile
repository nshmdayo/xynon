start:
	docker build -t nproxy -f ./build/Dockerfile .
	docker run --name nproxy -p 8000:8000 -it nproxy

stop:
	docker stop nproxy
	docker rm nproxy
	docker rmi nproxy

restart: stop start