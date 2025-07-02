local M = {}

local utils = require('mdtask.utils')
local ui = require('mdtask.ui')
local config = require('mdtask.config')

-- List all tasks
function M.list()
  local args = {'list', '--format', 'json'}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to list tasks: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local tasks = utils.parse_json(output)
    if #tasks == 0 then
      utils.notify('No tasks found')
      return
    end
    
    ui.show_task_list(tasks)
  end)
end

-- List tasks by status
function M.list_by_status(status)
  if not status or status == '' then
    status = 'todo'
  end
  
  local args = {'list', '--status', status, '--format', 'json'}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to list tasks: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local tasks = utils.parse_json(output)
    if #tasks == 0 then
      utils.notify('No ' .. status .. ' tasks found')
      return
    end
    
    ui.show_task_list(tasks, status:upper() .. ' Tasks')
  end)
end

-- Search tasks
function M.search(query)
  if not query or query == '' then
    vim.ui.input({ prompt = 'Search tasks: ' }, function(input)
      if input and input ~= '' then
        M.search(input)
      end
    end)
    return
  end
  
  local args = {'search', query, '--format', 'json'}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to search tasks: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local tasks = utils.parse_json(output)
    if #tasks == 0 then
      utils.notify('No tasks found for: ' .. query)
      return
    end
    
    ui.show_task_list(tasks, 'Search Results: ' .. query)
  end)
end

-- Create new task
function M.new()
  ui.show_task_form(function(task_data)
    local args = {'new'}
    
    -- Add title (required)
    if task_data.title and task_data.title ~= '' then
      table.insert(args, '--title')
      table.insert(args, task_data.title)
    else
      utils.notify('Title is required', vim.log.levels.ERROR)
      return
    end
    
    -- Add description (always provide this flag to avoid interactive prompts)
    table.insert(args, '--description')
    table.insert(args, task_data.description or '')
    
    -- Add content (always provide this flag to avoid interactive prompts)
    table.insert(args, '--content')
    table.insert(args, task_data.content or '')
    
    -- Add status
    if task_data.status and task_data.status ~= '' then
      table.insert(args, '--status')
      table.insert(args, task_data.status)
    end
    
    -- Add tags
    if task_data.tags and #task_data.tags > 0 then
      local valid_tags = {}
      for _, tag in ipairs(task_data.tags) do
        if tag and tag:match('^%s*(.-)%s*$') ~= '' then
          table.insert(valid_tags, tag:match('^%s*(.-)%s*$'))
        end
      end
      if #valid_tags > 0 then
        table.insert(args, '--tags')
        table.insert(args, table.concat(valid_tags, ','))
      end
    end
    
    utils.execute_mdtask(args, function(err, output)
      if err then
        utils.notify('Failed to create task: ' .. err, vim.log.levels.ERROR)
        return
      end
      
      utils.notify('Task created successfully')
      
      -- Refresh task list if it's open
      if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
        M.list()
      end
    end)
  end)
end

-- Edit task
function M.edit(task_id)
  if not task_id or task_id == '' then
    -- Get task ID from user or current line
    local current_line = vim.api.nvim_get_current_line()
    task_id = current_line:match('%(([^)]+)%)')
    
    if not task_id then
      vim.ui.input({ prompt = 'Task ID: ' }, function(input)
        if input and input ~= '' then
          M.edit(input)
        end
      end)
      return
    end
  end
  
  utils.get_task_by_id(task_id, function(err, task)
    if err then
      utils.notify('Failed to get task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    ui.show_task_form(function(task_data)
      local args = {'edit', task_id}
      
      -- Add updated fields
      if task_data.title then
        table.insert(args, '--title')
        table.insert(args, task_data.title)
      end
      
      if task_data.description then
        table.insert(args, '--description')
        table.insert(args, task_data.description)
      end
      
      if task_data.status then
        table.insert(args, '--status')
        table.insert(args, task_data.status)
      end
      
      if task_data.tags and #task_data.tags > 0 then
        table.insert(args, '--tags')
        table.insert(args, table.concat(task_data.tags, ','))
      end
      
      utils.execute_mdtask(args, function(err, output)
        if err then
          utils.notify('Failed to update task: ' .. err, vim.log.levels.ERROR)
          return
        end
        
        utils.notify('Task updated successfully')
        
        -- Refresh task list if it's open
        if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
          M.list()
        end
      end)
    end, task)
  end)
end

-- Archive task
function M.archive(task_id)
  if not task_id or task_id == '' then
    -- Get task ID from current line
    local current_line = vim.api.nvim_get_current_line()
    task_id = current_line:match('%(([^)]+)%)')
    
    if not task_id then
      vim.ui.input({ prompt = 'Task ID to archive: ' }, function(input)
        if input and input ~= '' then
          M.archive(input)
        end
      end)
      return
    end
  end
  
  local args = {'archive', task_id}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to archive task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    utils.notify('Task archived successfully')
    
    -- Refresh task list if it's open
    if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
      M.list()
    end
  end)
end

-- Open web interface
function M.open_web()
  local cfg = config.get()
  local args = {'web'}
  
  if cfg.web_port then
    table.insert(args, '--port')
    table.insert(args, tostring(cfg.web_port))
  end
  
  if not cfg.open_browser then
    table.insert(args, '--no-browser')
  end
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to start web server: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    utils.notify('Web server started')
  end)
end

return M