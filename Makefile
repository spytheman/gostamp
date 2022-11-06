VERSION=0.7
GOBUILD?=go build -ldflags "-X main.version=$(VERSION)"
.PHONY: clean binaries release

clean: 
	rm -rf binaries/ release/ gostamp

binaries: binaries/osx_x86_64/gostamp binaries/windows_x86_64/gostamp.exe binaries/linux_x86_64/gostamp binaries/windows_x86_32/gostamp.exe binaries/linux_x86_32/gostamp
	
binaries/osx_x86_64/gostamp:
	GOOS=darwin GOARCH=amd64  $(GOBUILD)  -o binaries/osx_x86_64/gostamp 

binaries/windows_x86_64/gostamp.exe:
	GOOS=windows GOARCH=amd64 $(GOBUILD)  -o binaries/windows_x86_64/gostamp.exe 

binaries/windows_x86_32/gostamp.exe:
	GOOS=windows GOARCH=386   $(GOBUILD)  -o binaries/windows_x86_32/gostamp.exe 

binaries/linux_x86_64/gostamp:
	GOOS=linux GOARCH=amd64   $(GOBUILD)  -o binaries/linux_x86_64/gostamp 

binaries/linux_x86_32/gostamp:
	GOOS=linux GOARCH=386     $(GOBUILD)  -o binaries/linux_x86_32/gostamp 

binaries/linux_arm64/gostamp:
	GOOS=linux GOARCH=arm64   $(GOBUILD)  -o binaries/linux_arm64/gostamp 

release: binaries
	mkdir -p release/ ;
	cd binaries/; for i in *; do cd $$i/ ; zip -r -D ../../release/gostamp_$(VERSION)_$$i.zip *; cd ../ ; done

format:
	gofmt -w main.go terminal/terminal.go

install:
	go install -ldflags "-X main.version=$(VERSION)"
