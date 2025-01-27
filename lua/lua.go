package lua

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type Lua struct {
	L  *lua.LState
	mx *sync.Mutex
}

func New() *Lua {
	L := lua.NewState()
	return &Lua{L: L}
}

func (l *Lua) Run(script string) error {
	return l.L.DoString(script)
}

func (l *Lua) Close() {
	l.L.Close()
}
