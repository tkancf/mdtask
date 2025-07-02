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