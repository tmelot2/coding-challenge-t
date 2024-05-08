.PHONY: buildSetup runSetup

setupName := "cctsetup"
setup:
	docker build -f ./Dockerfile-setup -t $(setupName) .
	docker run -it --rm --name $(setupName) $(setupName)
