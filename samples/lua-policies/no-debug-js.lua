-- Block console.log, debugger, and alert in JavaScript files
function check(file, content)
    if not string.match(file, "%.js$") and not string.match(file, "%.ts$") then
        return true, ""
    end

    if string.match(content, "console%.log") then
        return false, "Remove console.log from " .. file
    end
    if string.match(content, "debugger") then
        return false, "Remove debugger statement from " .. file
    end
    if string.match(content, "alert%(") then
        return false, "Remove alert() from " .. file
    end

    return true, ""
end
