local M = {}

local config = require('mdtask.config')
local tasks = require('mdtask.tasks')
local ui = require('mdtask.ui')

-- Setup function for the plugin
function M.setup(opts)
  config.setup(opts or {})
  
  -- Create user commands
  vim.api.nvim_create_user_command('MdTaskList', function()
    tasks.list()
  end, { desc = 'List all mdtask tasks' })
  
  vim.api.nvim_create_user_command('MdTaskNew', function()
    tasks.new()
  end, { desc = 'Create a new mdtask task' })
  
  vim.api.nvim_create_user_command('MdTaskSearch', function(opts)
    tasks.search(opts.args)
  end, { nargs = '*', desc = 'Search mdtask tasks' })
  
  vim.api.nvim_create_user_command('MdTaskStatus', function(opts)
    tasks.list_by_status(opts.args)
  end, { nargs = '?', desc = 'List tasks by status' })
  
  vim.api.nvim_create_user_command('MdTaskWeb', function()
    tasks.open_web()
  end, { desc = 'Open mdtask web interface' })
  
  vim.api.nvim_create_user_command('MdTaskEdit', function(opts)
    tasks.edit(opts.args)
  end, { nargs = '?', desc = 'Edit a task' })
  
  vim.api.nvim_create_user_command('MdTaskArchive', function(opts)
    tasks.archive(opts.args)
  end, { nargs = '?', desc = 'Archive a task' })
  
  vim.api.nvim_create_user_command('MdTaskDebug', function()
    local cfg = config.get()
    print('Config: ' .. vim.inspect(cfg))
    print('Current directory: ' .. vim.fn.getcwd())
    print('mdtask executable: ' .. vim.fn.executable('mdtask'))
    
    -- Test mdtask command
    local result = vim.fn.system('mdtask --help')
    print('mdtask help result: ' .. result)
  end, { desc = 'Debug mdtask configuration' })
  
  vim.api.nvim_create_user_command('MdTaskTestCreate', function()
    print('Testing direct command execution...')
    local result = vim.fn.system('mdtask new --title "test from debug" --status TODO')
    print('Direct command result: ' .. vim.inspect(result))
    print('Exit code: ' .. vim.v.shell_error)
    
    -- List files to see if anything was created
    local ls_result = vim.fn.system('ls -la *.md 2>/dev/null || echo "No .md files found"')
    print('Files: ' .. ls_result)
  end, { desc = 'Test mdtask new command directly' })
  
  -- Telescope integration if available
  if pcall(require, 'telescope') then
    require('mdtask.telescope').setup()
  end
end

-- Public API
M.list = tasks.list
M.new = tasks.new
M.search = tasks.search
M.edit = tasks.edit
M.archive = tasks.archive
M.open_web = tasks.open_web

return M