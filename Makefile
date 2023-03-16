test:
	go test ./...

build_gui:
	fyne-cross linux -arch arm -name gui ./cmd/gui/gui.go
