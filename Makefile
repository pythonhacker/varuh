
PROGRAM=varuh

all: program

program: scripts/main.go $(wildcard *.go)
	@echo "Building ${PROGRAM}"
	@go mod tidy
	@go build -o ${PROGRAM} scripts/main.go 

install: 
	@echo -n "Installing ${PROGRAM}"
	@if [ -f ./${PROGRAM} ]; then \
		sudo cp ./${PROGRAM} /usr/local/bin; \
		echo "...done"; \
	fi 

debian: program # Run as $ make debian VERSION=0.x.y
	@echo "Building debian package for version=${VERSION}"
	@mkdir -p debian/varuh-${VERSION}_amd64/usr/bin/
	@mkdir -p debian/varuh-${VERSION}_amd64/DEBIAN/
	@cp varuh debian/varuh-${VERSION}_amd64/usr/bin/
	@sed 's/VERSION/${VERSION}/g' META > debian/varuh-${VERSION}_amd64/DEBIAN/control
	@cd debian/ && dpkg-deb --build --root-owner-group varuh-${VERSION}_amd64/ && cd -
	@if [ -f debian/varuh-${VERSION}_amd64.deb ]; then \
		echo "Build successful."; \
	else \
		echo "Build failed."; \
	fi \

clean:
	@echo "Removing ${PROGRAM}"
	@rm -f ${PROGRAM}
