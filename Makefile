.PHONY: setup run

setupName := "queryToolDbSetup"
setup:
	docker build -f ./Dockerfile-setup -t $(setupName) .
	docker run -it --rm --name $(setupName) $(setupName)

run:
	go run main.go queryTool.go
