-- Block hardcoded secrets and API keys
function check(file, content)
    local patterns = {
        {"AKIA[A-Z0-9]{16}", "AWS Access Key"},
        {"ghp_[A-Za-z0-9]{36}", "GitHub Personal Access Token"},
        {"sk%-[A-Za-z0-9]{48}", "OpenAI API Key"},
        {"xox[baprs]%-[A-Za-z0-9%-]+", "Slack Token"},
        {"-----BEGIN PRIVATE KEY-----", "Private Key"},
        {"-----BEGIN RSA PRIVATE KEY", "RSA Private Key"},
    }

    for _, p in ipairs(patterns) do
        if string.match(content, p[1]) then
            return false, p[2] .. " detected in " .. file .. " - remove before committing"
        end
    end

    return true, ""
end
