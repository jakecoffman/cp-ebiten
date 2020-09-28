module github.com/jakecoffman/cpebiten

go 1.15

require (
	github.com/fogleman/gg v1.3.0
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20200707082815-5321531c36a2 // indirect
	github.com/hajimehoshi/ebiten v1.11.8
	github.com/jakecoffman/cp v1.0.0
	golang.org/x/exp v0.0.0-20200924195034-c827fd4f18b9 // indirect
	golang.org/x/image v0.0.0-20200927005634-a67d67e0935b // indirect
	golang.org/x/mobile v0.0.0-20200801112145-973feb4309de // indirect
	golang.org/x/sys v0.0.0-20200926100807-9d91bd62050c // indirect
)

// my fork is single threaded and increases performance by ~3x for windows and osx
replace github.com/hajimehoshi/ebiten => github.com/jakecoffman/ebiten v1.13.0-alpha.0.20200926172331-1d10970a8c48
