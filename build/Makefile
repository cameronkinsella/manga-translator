build-all: windows-amd64 windows-386 linux-amd64 linux-386 darwin-amd64 darwin-arm64

windows-amd64:
	GOOS=windows GOARCH=amd64 go build -v -o windows-amd64/ -ldflags -H=windowsgui ../cmd/manga-translator
	GOOS=windows GOARCH=amd64 go build -v -o windows-amd64/ ../cmd/manga-translator-setup

windows-386:
	GOOS=windows GOARCH=386 go build -v -o windows-386/ -ldflags -H=windowsgui ../cmd/manga-translator
	GOOS=windows GOARCH=386 go build -v -o windows-386/ ../cmd/manga-translator-setup

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -v -o linux-amd64/ ../cmd/manga-translator
	GOOS=linux GOARCH=amd64 go build -v -o linux-amd64/ ../cmd/manga-translator-setup

linux-386:
	GOOS=linux GOARCH=386 go build -v -o linux-386/ ../cmd/manga-translator
	GOOS=linux GOARCH=386 go build -v -o linux-386/ ../cmd/manga-translator-setup

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -v -o darwin-amd64/ ../cmd/manga-translator
	GOOS=darwin GOARCH=amd64 go build -v -o darwin-amd64/ ../cmd/manga-translator-setup

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -v -o darwin-arm64/ ../cmd/manga-translator
	GOOS=darwin GOARCH=aarm64 go build -v -o darwin-arm64/ ../cmd/manga-translator-setup
