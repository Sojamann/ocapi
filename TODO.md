# Features
[ ] Image diff
[ ] Registry graph
[ ] Base swap
[ ] other config formats besides docker

# Perf / ??
[ ] we must limit the amount of requests that go out!!
    if we fetch all images/manifests than we make 1000+ requests
    the server does not like this

# Usability
[ ] good tui design
[ ] clear error handling


# Libs that might be of use
Futures, Result, Optional, ....
 -> https://pkg.go.dev/github.com/samber/mo@v1.8.0

Functional programming, ....
 -> https://pkg.go.dev/github.com/samber/lo@v1.37.0

TUI
 -> ps://github.com/charmbracelet/bubbletea
 -> ps://github.com/charmbracelet/lipgloss
 -> ps://github.com/charmbracelet/bubbles
