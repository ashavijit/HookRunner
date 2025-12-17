local COPYRIGHT = "Copyright"

function check(file, content)
    local extensions = {"%.go$", "%.js$", "%.ts$", "%.py$", "%.java$", "%.c$", "%.cpp$"}
    local is_source = false

    for _, ext in ipairs(extensions) do
        if string.match(file, ext) then
            is_source = true
            break
        end
    end

    if not is_source then
        return true, ""
    end

    local lines = 0
    for line in string.gmatch(content, "[^\n]+") do
        lines = lines + 1
        if string.match(line, COPYRIGHT) then
            return true, ""
        end
        if lines > 10 then
            break
        end
    end

    return false, "Missing copyright header in " .. file
end
