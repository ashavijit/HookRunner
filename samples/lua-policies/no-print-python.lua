-- Block Python print statements (use logging instead)
function check(file, content)
    if not string.match(file, "%.py$") then
        return true, ""
    end

    if string.match(content, "print%(") then
        return false, "Use logging instead of print() in " .. file
    end

    return true, ""
end
