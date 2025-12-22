# Clean-Room Architecture

Clean-room mode provides CI parity by running hooks in an isolated temporary directory containing only staged files.

## Problem

Local development environments contain:
- Unstaged changes
- Untracked files
- Build artifacts
- IDE-generated files

Hooks running locally may pass due to these extra files, then fail in CI where only committed code exists.

## Solution

The `--clean-room` flag creates an isolated environment:

```
hookrunner run pre-commit --clean-room
```

## How It Works

```
1. Create temp directory
   └── /tmp/hookrunner-clean-XXXXX/

2. Export staged files via git
   └── git archive --format=tar HEAD | tar -x -C /tmp/...

3. Copy staged changes
   └── For each staged file: copy to temp dir

4. Run hooks in temp directory
   └── All hooks execute with workDir = temp dir

5. Cleanup
   └── Remove temp directory on exit
```

## Implementation

### File: internal/git/cleanroom.go

```go
func CreateCleanRoom() (string, error) {
    // Create temp directory
    tmpDir, err := os.MkdirTemp("", "hookrunner-clean-")

    // Get repository root
    root := GetRepoRoot()

    // Export HEAD to temp directory
    cmd := exec.Command("git", "archive", "--format=tar", "HEAD")
    // Pipe to tar extraction in tmpDir

    // Copy staged files (overwrite with staged versions)
    stagedFiles := GetStagedFiles()
    for _, file := range stagedFiles {
        // Read staged content: git show :file
        // Write to tmpDir/file
    }

    return tmpDir, nil
}
```

### CLI Integration

```go
// In runHook function
if cleanRoom {
    cleanRoomDir, err := git.CreateCleanRoom()
    defer git.CleanupCleanRoom(cleanRoomDir)
    executionDir = cleanRoomDir
}
```

## Usage

```bash
# Run with clean-room (prompts for confirmation)
hookrunner run pre-commit --clean-room

# Output:
# Clean-room mode: Hooks will run in an isolated temporary directory
# Warning: This excludes all unstaged changes and untracked files.
# Proceed with clean-room execution? [y/N]: y
# Running hooks in: /tmp/hookrunner-clean-123456/
```

## When to Use

| Scenario | Use Clean-Room |
|----------|----------------|
| Debugging CI failures | Yes |
| Pre-push validation | Yes |
| Quick local iteration | No |
| Testing uncommitted work | No |

## Limitations

1. Requires git repository
2. Only includes tracked files
3. Slower than normal mode (file copying overhead)
4. Prompts for confirmation (cannot be automated without flag)
