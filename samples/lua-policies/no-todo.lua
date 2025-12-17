-- Block files containing TODO or FIXME comments
function check(file, content)
    if string.match(content, "TODO") then
        return false, "Remove TODO comments before commit: " .. file
    end
    if string.match(content, "FIXME") then
        return false, "Remove FIXME comments before commit: " .. file
    end
    return true, ""
end
