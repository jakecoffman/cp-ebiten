module github.com/jakecoffman/cpebiten

go 1.15

require (
	github.com/fogleman/gg v1.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/hajimehoshi/ebiten/v2 v2.1.0-alpha.0.20201017083052-1d82aec7129e
	github.com/jakecoffman/cp v1.0.0
	golang.org/x/exp v0.0.0-20201008143054-e3b2a7f2fdc7 // indirect
	golang.org/x/image v0.0.0-20200927104501-e162460cd6b5
	golang.org/x/sys v0.0.0-20201016160150-f659759dc4ca // indirect
)

replace github.com/hajimehoshi/ebiten/v2 => github.com/jakecoffman/ebiten/v2 v2.1.0-alpha.0.20201018160528-0018f37928ca
