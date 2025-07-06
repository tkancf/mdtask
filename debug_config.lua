-- Debug script to test mdtask binary detection
local function find_mdtask_binary()
  -- First try current working directory (prefer local development version)
  local cwd = vim.fn.getcwd()
  print("Current working directory:", cwd)
  
  local local_binary = cwd .. '/mdtask'
  print("Checking local_binary:", local_binary)
  print("Executable?", vim.fn.executable(local_binary))
  if vim.fn.executable(local_binary) == 1 then
    print("Found local binary:", local_binary)
    return local_binary
  end
  
  -- Try relative path
  print("Checking ./mdtask")
  print("Executable?", vim.fn.executable('./mdtask'))
  if vim.fn.executable('./mdtask') == 1 then
    print("Found relative binary: ./mdtask")
    return './mdtask'
  end
  
  -- Try parent directory (in case we're in a subdirectory)
  local parent_binary = cwd .. '/../mdtask'
  print("Checking parent_binary:", parent_binary)
  print("Executable?", vim.fn.executable(parent_binary))
  if vim.fn.executable(parent_binary) == 1 then
    print("Found parent binary:", parent_binary)
    return parent_binary
  end
  
  -- Try looking for mdtask project root by finding git root
  local git_root = vim.fn.system('git rev-parse --show-toplevel 2>/dev/null'):gsub('\n', '')
  print("Git root:", git_root)
  if git_root and git_root ~= '' then
    local git_root_binary = git_root .. '/mdtask'
    print("Checking git_root_binary:", git_root_binary)
    print("Executable?", vim.fn.executable(git_root_binary))
    if vim.fn.executable(git_root_binary) == 1 then
      print("Found git root binary:", git_root_binary)
      return git_root_binary
    end
  end
  
  -- Fall back to system PATH
  print("Falling back to system PATH")
  return 'mdtask'
end

local result = find_mdtask_binary()
print("Final result:", result)

-- Force reload the config with the correct path
package.loaded['mdtask.config'] = nil
local config = require('mdtask.config')
config.setup({mdtask_path = result})
print("New config mdtask_path:", config.get().mdtask_path)