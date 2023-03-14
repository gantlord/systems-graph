find . -name '*.go' | while read gofile; do gofmt -w $gofile; done
