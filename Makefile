# note: call scripts from /scripts

GOCMD=go
TAG=release-0_1

all: setup check

setup:
	$(GOCMD) mod tidy

check: static_check
	@echo "check done"

static_check:
	golangci-lint run --fix

image: stargate.o dimension.o
	@echo "build done"

stargate.o:
	docker build --force-rm -t wormhole-stargate:${TAG} -f cmd/stargate/Dockerfile .
	docker save wormhole-stargate:${TAG} -o build/package/stargate-${TAG}.tar
dimension.o:
	docker build --force-rm -t wormhole-dimension:${TAG} -f cmd/dimension/Dockerfile .
	docker save wormhole-dimension:${TAG} -o build/package/dimension-${TAG}.tar

clean:
	rm -rf build/package/dimension-* build/package/stargate-*
	docker rmi wormhole-stargate:${TAG}
	docker rmi wormhole-dimension:${TAG}