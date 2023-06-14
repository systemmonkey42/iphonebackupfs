
install:
	go build --tags osusergo,netgo -o iphonefs -ldflags "-w -s" .
	mv iphonefs /usr/local/bin
