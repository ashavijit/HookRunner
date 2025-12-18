# Lua Policy Samples

Sample Lua policies for HookRunner. Add to your `hooks.yaml`:

```yaml
policies:
  lua_scripts:
    - samples/lua-policies/no-secrets.lua
    - samples/lua-policies/no-todo.lua
```

## Available Policies

| Policy | Description |
|--------|-------------|
| `no-secrets.lua` | Block AWS keys, GitHub tokens, private keys |
| `no-todo.lua` | Block TODO/FIXME comments |
| `no-debug-js.lua` | Block console.log, debugger, alert in JS/TS |
| `no-print-python.lua` | Block print() in Python |
| `max-file-size.lua` | Enforce 100KB file size limit |
| `require-copyright.lua` | Require copyright header |

## Writing Custom Policies

```lua
function check(file, content)
    if string.match(content, "bad_pattern") then
        return false, "Error message"
    end
    return true, ""
end
```
