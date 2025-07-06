local M = {}

local config = require('mdtask.config')

-- Execute mdtask command and return result
function M.execute_mdtask(args, callback, stdin_input, skip_json)
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
  
  -- Add JSON format flag by default for commands that support it
  if not skip_json and (args[1] == 'list' or args[1] == 'search' or args[1] == 'get') then
    table.insert(full_args, '--format')
    table.insert(full_args, 'json')
  end
  
  -- Add path arguments if configured
  if cfg.task_paths and #cfg.task_paths > 0 then
    for _, path in ipairs(cfg.task_paths) do
      table.insert(full_args, '--paths')
      table.insert(full_args, path)
    end
  end
  
  
  if callback then
    -- Async execution
    local stdout_data = {}
    local stderr_data = {}
    
    local job_opts = {
      stdout_buffered = true,
      stderr_buffered = true,
      cwd = vim.fn.getcwd(),
      timeout = 10000,  -- 10 second timeout
      on_stdout = function(_, data)
        if data and #data > 0 then
          for _, line in ipairs(data) do
            if line ~= '' then
              table.insert(stdout_data, line)
            end
          end
        end
      end,
      on_stderr = function(_, data)
        if data and #data > 0 then
          for _, line in ipairs(data) do
            if line ~= '' then
              table.insert(stderr_data, line)
            end
          end
        end
      end,
      on_exit = function(_, code)
        local stdout_output = table.concat(stdout_data, '\n')
        local stderr_output = table.concat(stderr_data, '\n')
        
        -- For mdtask new command, success is determined by exit code
        -- Even if there's stderr output (like interactive prompts), it might still succeed
        if code == 0 then
          callback(nil, stdout_output)
        else
          -- Command failed
          local error_msg = stderr_output
          if error_msg == '' then
            error_msg = 'Command failed with exit code: ' .. code
          end
          callback(error_msg, nil)
        end
      end
    }
    
    local job_id = vim.fn.jobstart({cmd, unpack(full_args)}, job_opts)
    
    if job_id == 0 then
      callback('Failed to start job', nil)
    elseif job_id == -1 then
      callback('Invalid command', nil)
    elseif stdin_input then
      -- Send stdin input to the job
      vim.fn.chansend(job_id, stdin_input)
      vim.fn.chanclose(job_id, 'stdin')
    end
  else
    -- Sync execution
    if stdin_input then
      -- Use system with input
      local result = vim.fn.system({cmd, unpack(full_args)}, stdin_input)
      if vim.v.shell_error ~= 0 then
        vim.notify('mdtask command failed: ' .. result, vim.log.levels.ERROR)
        return nil
      end
      return result
    else
      local result = vim.fn.system({cmd, unpack(full_args)})
      if vim.v.shell_error ~= 0 then
        vim.notify('mdtask command failed: ' .. result, vim.log.levels.ERROR)
        return nil
      end
      return result
    end
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

-- Format task for display with view mode support
function M.format_task(task, view_mode)
  local status = task.status or 'TODO'
  local title = task.title or 'Untitled'
  local id = task.id or ''
  local description = task.description
  view_mode = view_mode or 'detailed'
  
  -- Debug: check task structure
  -- vim.notify('Task ID: ' .. tostring(task.id) .. ', Status: ' .. tostring(task.status), vim.log.levels.INFO)
  
  -- Generate file path from task ID
  local file_path = ''
  if id and id ~= '' then
    local timestamp = id:match('task/(.+)')
    if timestamp then
      file_path = timestamp .. '.md'
    end
  end
  
  -- Check deadline status
  local deadline_status = nil
  if task.deadline then
    local deadline_date = vim.fn.strptime('%Y-%m-%dT%H:%M:%SZ', task.deadline)
    if deadline_date and deadline_date < os.time() then
      deadline_status = 'overdue'
    else
      deadline_status = 'due'
    end
  end
  
  -- Add visual indicators (text-based)
  local indicators = {}
  
  -- Priority indicator
  if task.tags then
    for _, tag in ipairs(task.tags) do
      if tag:match('priority/high') or tag:match('urgent') then
        table.insert(indicators, '[!]')
        break
      end
    end
  end
  
  -- Deadline urgency indicators
  if task.deadline then
    local today = os.date('%Y-%m-%d')
    local deadline_date = task.deadline:sub(1, 10) -- Get just the date part
    
    if deadline_date < today then
      table.insert(indicators, '[OVERDUE]') 
    elseif deadline_date == today then
      table.insert(indicators, '[TODAY]')
    else
      -- Check if due within next 3 days
      local deadline_time = os.time{year=deadline_date:sub(1,4), month=deadline_date:sub(6,7), day=deadline_date:sub(9,10)}
      local today_time = os.time()
      local three_days = today_time + (3 * 24 * 60 * 60)
      if deadline_time <= three_days then
        table.insert(indicators, '[SOON]')
      end
    end
  end
  
  -- Build indicator string
  local indicator_str = ''
  if #indicators > 0 then
    indicator_str = table.concat(indicators, '') .. ' '
  end
  
  -- Format main line with indicators
  local lines = {}
  local main_line = string.format('- %s%s: %s', indicator_str, status, title)
  table.insert(lines, main_line)
  
  -- Compact mode: only show main line
  if view_mode == 'compact' then
    return lines, deadline_status, id
  end
  
  -- Detailed mode: show all information
  -- Add markdown link as second line if file path exists
  if file_path ~= '' then
    table.insert(lines, string.format('    - [%s](%s)', title, file_path))
  end
  
  -- Add description as third line if present
  if description and description ~= '' then
    table.insert(lines, string.format('    - %s', description))
  end
  
  -- Add deadline date as additional line if present
  if task.deadline then
    -- Parse deadline and format as readable date
    local year, month, day = task.deadline:match('(%d%d%d%d)%-(%d%d)%-(%d%d)')
    if year and month and day then
      local deadline_text = string.format('    - Deadline: %s/%s/%s', year, month, day)
      table.insert(lines, deadline_text)
    end
  end
  
  -- Add parent task info if present
  if task.parent_id and task.parent_id ~= '' then
    local parent_id_short = task.parent_id:match('task/(.+)') or task.parent_id
    table.insert(lines, string.format('    - Parent: %s', parent_id_short))
  end
  
  -- Return lines, deadline status, and task ID separately
  return lines, deadline_status, id
end

-- Get task ID from current buffer file path
function M.get_task_id_from_buffer()
  local buf = vim.api.nvim_get_current_buf()
  local file_path = vim.api.nvim_buf_get_name(buf)
  
  if not file_path or file_path == '' then
    return nil
  end
  
  -- Extract filename without extension
  local filename = file_path:match('[^/]+$')
  if not filename or not filename:match('%.md$') then
    return nil
  end
  
  -- Extract timestamp from filename (format: YYYYMMDDHHMMSS.md)
  local timestamp = filename:match('^(%d%d%d%d%d%d%d%d%d%d%d%d%d%d)%.md$')
  if timestamp then
    return 'task/' .. timestamp
  end
  
  return nil
end

-- Get task by ID
function M.get_task_by_id(task_id, callback)
  if not task_id or task_id == '' then
    if callback then callback('Task ID is required', nil) end
    return nil
  end
  
  local args = {'get', task_id}
  
  M.execute_mdtask(args, function(err, output)
    if err then
      if callback then callback(err, nil) end
      return
    end
    
    local task = M.parse_json(output)
    if task then
      if callback then callback(nil, task) end
    else
      if callback then callback('Task not found: ' .. task_id, nil) end
    end
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