cmd /C "set GOOS=linux&& set GOARCH=amd64&& go build -o builds/disneytables-linux-amd64"
cmd /C "set GOOS=windows&& set GOARCH=amd64&& go build -o builds/disneytables-win-amd64.exe"
