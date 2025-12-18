package lua

import (
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

type PolicyResult struct {
	Passed  bool
	Message string
	File    string
	Line    int
}

type Runner struct {
	workDir string
}

func NewRunner(workDir string) *Runner {
	return &Runner{workDir: workDir}
}

func (r *Runner) RunPolicy(scriptPath string, files []string) ([]PolicyResult, error) {
	L := lua.NewState()
	defer L.Close()

	var results []PolicyResult

	L.SetGlobal("block", L.NewFunction(func(L *lua.LState) int {
		msg := L.OptString(1, "Policy violation")
		file := L.OptString(2, "")
		results = append(results, PolicyResult{
			Passed:  false,
			Message: msg,
			File:    file,
		})
		return 0
	}))

	L.SetGlobal("pass", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))

	L.SetGlobal("read_file", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		fullPath := filepath.Join(r.workDir, path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(string(data)))
		return 1
	}))

	L.SetGlobal("match", L.NewFunction(func(L *lua.LState) int {
		str := L.CheckString(1)
		pattern := L.CheckString(2)
		matched, _ := filepath.Match(pattern, str)
		L.Push(lua.LBool(matched))
		return 1
	}))

	filesTable := L.NewTable()
	for _, f := range files {
		filesTable.Append(lua.LString(f))
	}
	L.SetGlobal("files", filesTable)
	L.SetGlobal("workdir", lua.LString(r.workDir))

	if err := L.DoFile(scriptPath); err != nil {
		return nil, fmt.Errorf("lua script error: %w", err)
	}

	checkFn := L.GetGlobal("check")
	if checkFn.Type() == lua.LTFunction {
		for _, file := range files {
			content, _ := os.ReadFile(filepath.Join(r.workDir, file))

			if err := L.CallByParam(lua.P{
				Fn:      checkFn,
				NRet:    2,
				Protect: true,
			}, lua.LString(file), lua.LString(string(content))); err != nil {
				return nil, err
			}

			passed := L.ToBool(-2)
			msg := L.ToString(-1)
			L.Pop(2)

			if !passed {
				results = append(results, PolicyResult{
					Passed:  false,
					Message: msg,
					File:    file,
				})
			}
		}
	}

	return results, nil
}

func (r *Runner) RunScript(scriptPath string) error {
	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("workdir", lua.LString(r.workDir))
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		fmt.Println(msg)
		return 0
	}))

	return L.DoFile(scriptPath)
}
