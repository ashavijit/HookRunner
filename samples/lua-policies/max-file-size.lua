-- Enforce max file size (100KB)
local MAX_SIZE = 100 * 1024

function check(file, content)
    if #content > MAX_SIZE then
        local size_kb = math.floor(#content / 1024)
        return false, file .. " is too large (" .. size_kb .. "KB > 100KB limit)"
    end
    return true, ""
end
