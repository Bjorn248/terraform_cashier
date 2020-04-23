build:
	go build -o cashier main.go

release:
	gox -osarch="linux/amd64 darwin/amd64 windows/amd64"
	tar cfvz cashier_osx.tar.gz terraform_cashier_darwin_amd64
	tar cfvz cashier_windows.tar.gz terraform_cashier_windows_amd64.exe
	tar cfvz cashier_linux.tar.gz terraform_cashier_linux_amd64
	rm terraform_cashier_*

clean:
	rm cashier*

test:
	go test -v ./...
