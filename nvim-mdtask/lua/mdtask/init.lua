local M = {}

local config = require('mdtask.config')
local tasks = require('mdtask.tasks')
local ui = require('mdtask.ui')
local highlights = require('mdtask.highlights')

-- Setup function for the plugin
function M.setup(opts)
  config.setup(opts or {})
  highlights.setup()
  
  -- Create main command with subcommands
  vim.api.nvim_create_user_command('MdTask', function(opts)
    local subcommand = opts.fargs[1]
    local args = vim.list_slice(opts.fargs, 2)
    
    if not subcommand then
      tasks.list()
      return
    end
    
    local subcommands = {
      list = function()
        if args[1] then
          -- If argument provided, use it as status filter
          tasks.list_by_status(args[1])
        else
          tasks.list()
        end
      end,
      new = tasks.new,
      search = function()
        tasks.search(table.concat(args, ' '))
      end,
      edit = function()
        -- Check if editing specific field
        if args[1] == 'status' or args[1] == 'title' or args[1] == 'description' then
          -- :MdTask edit status [task_id] [value]
          local field = args[1]
          local task_id = args[2]
          local value = args[3]
          tasks.edit_field(task_id, field, value)
        else
          -- :MdTask edit [task_id]
          tasks.edit(args[1])
        end
      end,
      archive = function()
        tasks.archive(args[1])
      end,
      web = tasks.open_web,
      toggle = function()
        local actions = require('mdtask.actions')
        actions.toggle_task_status(args[1])
      end,
      preview = function()
        local actions = require('mdtask.actions')
        actions.preview_task(args[1])
      end,
      help = function()
        local help_text = [[
MdTask - Task management commands

Usage: :MdTask <subcommand> [args]

Subcommands:
  list [status]    List tasks (optionally filtered by status)
  new              Create a new task
  search <query>   Search tasks
  edit <id>        Edit a task (full form)
  edit <field> [id] [value]  Edit specific field (status/title/description)
  archive <id>     Archive a task
  web              Open web interface
  toggle <id>      Toggle task status
  preview <id>     Preview task details
  help             Show this help

Examples:
  :MdTask                    List all tasks
  :MdTask list               List all tasks
  :MdTask list TODO          List TODO tasks
  :MdTask new                Create new task
  :MdTask search bug fix     Search for "bug fix"
  :MdTask edit task/123      Edit task (full form)
  :MdTask edit status        Edit status of current task
  :MdTask edit status task/123  Edit status of specific task
  :MdTask edit title         Edit title of current task
  :MdTask toggle task/123    Toggle task status]]
        
        vim.notify(help_text, vim.log.levels.INFO)
      end,
    }
    
    local handler = subcommands[subcommand]
    if handler then
      handler()
    else
      vim.notify('Unknown subcommand: ' .. subcommand .. '\nUse :MdTask help for available commands', vim.log.levels.ERROR)
    end
  end, {
    nargs = '*',
    desc = 'MdTask commands',
    complete = function(ArgLead, CmdLine, CursorPos)
      local parts = vim.split(CmdLine, '%s+')
      
      -- Complete subcommands
      if #parts == 2 then
        local subcommands = {'list', 'new', 'search', 'edit', 'archive', 'web', 'toggle', 'preview', 'help'}
        return vim.tbl_filter(function(cmd)
          return cmd:find('^' .. ArgLead)
        end, subcommands)
      end
      
      -- Complete for edit subcommand
      if #parts == 3 and parts[2] == 'edit' then
        -- First argument after edit can be task ID or field name
        local fields = {'status', 'title', 'description'}
        return vim.tbl_filter(function(field)
          return field:find('^' .. ArgLead:lower())
        end, fields)
      end
      
      -- Complete status values
      if #parts == 3 and parts[2] == 'list' then
        local statuses = {'TODO', 'WIP', 'WAIT', 'SCHE', 'DONE'}
        return vim.tbl_filter(function(status)
          return status:find('^' .. ArgLead:upper())
        end, statuses)
      end
      
      -- Complete status values for edit status
      if #parts == 5 and parts[2] == 'edit' and parts[3] == 'status' then
        local statuses = {'TODO', 'WIP', 'WAIT', 'SCHE', 'DONE'}
        return vim.tbl_filter(function(status)
          return status:find('^' .. ArgLead:upper())
        end, statuses)
      end
      
      return {}
    end,
  })
  
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

-- Actions API
local actions = require('mdtask.actions')
M.toggle_status = actions.toggle_task_status
M.quick_archive = actions.quick_archive
M.preview = actions.preview_task

return M