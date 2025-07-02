local M = {}

local config = require('mdtask.config')

-- Execute mdtask command and return result
function M.execute_mdtask(args, callback)
  local cfg = config.get()
  local cmd = cfg.mdtask_path
  local full_args = {}
  
  -- Add command arguments first
  if type(args) == 'string' then
    table.insert(full_args, args)
  elseif type(args) == 'table' then
    for _, arg in ipairs(args) do
      table.insert(full_args, arg)
    end
  end
  
  -- Add path arguments if configured
  if cfg.task_paths and #cfg.task_paths > 0 then
    for _, path in ipairs(cfg.task_paths) do
      table.insert(full_args, '--paths')
      table.insert(full_args, path)
    end
  end
  
  -- Debug output
  local cmd_str = cmd .. ' ' .. table.concat(full_args, ' ')
  print('Executing: ' .. cmd_str)
  
  if callback then
    -- Async execution
    local job_id = vim.fn.jobstart({cmd, unpack(full_args)}, {
      stdout_buffered = true,
      stderr_buffered = true,
      cwd = vim.fn.getcwd(),
      on_stdout = function(_, data)
        if data and #data > 0 then
          local output = table.concat(data, '\n')
          print('stdout: ' .. output)
          callback(nil, output)
        end
      end,
      on_stderr = function(_, data)
        if data and #data > 0 then
          local error_output = table.concat(data, '\n')
          print('stderr: ' .. error_output)
          callback(error_output, nil)
        end
      end,
      on_exit = function(_, code)
        print('Exit code: ' .. code)
        if code ~= 0 then
          callback('Command failed with exit code: ' .. code, nil)
        else
          -- If no stdout was captured but command succeeded
          callback(nil, '')
        end
      end
    })
    
    if job_id == 0 then
      callback('Failed to start job', nil)
    elseif job_id == -1 then
      callback('Invalid command', nil)
    end
  else
    -- Sync execution
    local result = vim.fn.system({cmd, unpack(full_args)})
    if vim.v.shell_error ~= 0 then
      vim.notify('mdtask command failed: ' .. result, vim.log.levels.ERROR)
      return nil
    end
    return result
  end
end

-- Parse JSON output from mdtask
function M.parse_json(json_str)
  if not json_str or json_str == '' then
    return {}
  end
  
  local ok, result = pcall(vim.fn.json_decode, json_str)
  if not ok then
    vim.notify('Failed to parse mdtask JSON output', vim.log.levels.ERROR)
    return {}
  end
  
  return result
end

-- Format task for display
function M.format_task(task)
  local status_icon = {
    TODO = '○',
    WIP = '◐',
    WAIT = '◔',
    DONE = '●',
  }
  
  local icon = status_icon[task.status] or '?'
  local title = task.title or 'Untitled'
  local id = task.id or ''
  
  return string.format('%s %s (%s)', icon, title, id)
end

-- Get task by ID
function M.get_task_by_id(task_id, callback)
  if not task_id or task_id == '' then
    if callback then callback('Task ID is required', nil) end
    return nil
  end
  
  local args = {'list', '--format', 'json', '--all'}
  
  M.execute_mdtask(args, function(err, output)
    if err then
      if callback then callback(err, nil) end
      return
    end
    
    local tasks = M.parse_json(output)
    for _, task in ipairs(tasks) do
      if task.id == task_id then
        if callback then callback(nil, task) end
        return
      end
    end
    
    if callback then callback('Task not found: ' .. task_id, nil) end
  end)
end

-- Create floating window
function M.create_float_win(opts)
  local cfg = config.get()
  opts = opts or {}
  
  local width = opts.width or cfg.ui.width
  local height = opts.height or cfg.ui.height
  local border = opts.border or cfg.ui.border
  
  local win_width = vim.api.nvim_get_option('columns')
  local win_height = vim.api.nvim_get_option('lines')
  
  local row = math.floor((win_height - height) / 2)
  local col = math.floor((win_width - width) / 2)
  
  local buf = vim.api.nvim_create_buf(false, true)
  local win = vim.api.nvim_open_win(buf, true, {
    relative = 'editor',
    width = width,
    height = height,
    row = row,
    col = col,
    border = border,
    style = 'minimal',
  })
  
  return buf, win
end

-- Show notification
function M.notify(msg, level)
  level = level or vim.log.levels.INFO
  vim.notify('[mdtask] ' .. msg, level)
end

return M