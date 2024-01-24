module github.com/raff/gio-games/etris

go 1.20

require (
	github.com/hajimehoshi/ebiten/v2 v2.6.3
	github.com/raff/ebi-games/util v0.0.0-20230926050036-25f2223a5faf
)

require (
	github.com/ebitengine/purego v0.5.0 // indirect
	github.com/jezek/xgb v1.1.0 // indirect
	golang.org/x/exp/shiny v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/image v0.12.0 // indirect
	golang.org/x/mobile v0.0.0-20230922142353-e2f452493d57 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

replace github.com/raff/ebi-games/util v0.0.0-20230926050036-25f2223a5faf => ../util
