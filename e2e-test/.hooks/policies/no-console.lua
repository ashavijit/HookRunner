-- check if files contain console.log
function check(file, content)
    if string.match(file, "%.js$") and string.match(content, "console%.log") then
        return false, "Remove console.log before commit"
    end

    if string.match(file, "%.yaml$") and string.match(content, "password:") then
        return false, "Hardcoded password in YAML file"
    end

    return true, ""
end
