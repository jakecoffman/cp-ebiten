# cp-ebiten

Physics examples in [Ebiten](https://github.com/hajimehoshi/ebiten) using the [Go Chipmunk2D port](https://github.com/jakecoffman/cp).

## building WASM

`GOOS=js GOARCH=wasm go build -o tumble/tumble.wasm github.com/jakecoffman/tumble`
