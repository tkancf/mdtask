local M = {}

-- Function to find mdtask binary
local function find_mdtask_binary()
  -- First try current working directory (prefer local development version)
  local cwd = vim.fn.getcwd()
  local local_binary = cwd .. '/mdtask'
  if vim.fn.executable(local_binary) == 1 then
    return local_binary
  end
  
  -- Try relative path
  if vim.fn.executable('./mdtask') == 1 then
    return './mdtask'
  end
  
  -- Fall back to system PATH
  return 'mdtask'
end

-- Default configuration
local defaults = {
  -- Path to mdtask binary
  mdtask_path = find_mdtask_binary(),
  
  -- Default paths to search for tasks
  task_paths = {},  -- Empty means use current directory
  
  -- Web server port
  web_port = 7000,
  
  -- Open browser when starting web server
  open_browser = true,
  
  -- Telescope configuration
  telescope = {
    -- Use telescope for task selection
    enabled = true,
    -- Telescope theme
    theme = 'dropdown',
    -- Show task preview
    show_preview = true,
  },
  
  -- UI configuration
  ui = {
    -- Window width for floating windows
    width = 80,
    -- Window height for floating windows
    height = 20,
    -- Border style
    border = 'rounded',
  },
  
  -- Task creation defaults
  task_defaults = {
    -- Default status for new tasks
    status = 'TODO',
    -- Default tags to add to new tasks
    tags = {},
  },
}

M.options = {}

function M.setup(opts)
  M.options = vim.tbl_deep_extend('force', defaults, opts or {})
end

function M.get()
  return M.options
end

-- Reload configuration (useful for development)
function M.reload()
  M.options = vim.tbl_deep_extend('force', defaults, {})
  return M.options
end

return M