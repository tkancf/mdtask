local M = {}

local utils = require('mdtask.utils')

-- Parse a task line and extract information
function M.parse_task_line(line)
  -- First, try to find task ID in curly braces
  local task_id = line:match('{(task/%d+)}')
  if not task_id then
    return {}
  end
  
  -- Pattern: - STATUS: Title
  local status, title = line:match('^%s*%- (%w+): (.+)$')
  
  if status and title then
    -- Clean title by removing task ID in curly braces and trimming
    title = title:gsub('%s*{.+}%s*$', ''):match('^%s*(.-)%s*$')
  end
  
  return {
    status = status,
    title = title,
    task_id = task_id
  }
end

-- Parse description line
function M.parse_description_line(line)
  -- Pattern:     - description
  local description = line:match('^%s+%- (.+)$')
  -- Skip if it's a markdown link line
  if description and description:match('^%[.+%]%(.*%)$') then
    return nil
  end
  return description
end

-- Parse the entire buffer and extract task updates
function M.parse_buffer_for_updates(buf)
  local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
  local updates = {}
  local current_task = nil
  
  for i, line in ipairs(lines) do
    -- Skip header and separator lines
    if i <= 3 then
      goto continue
    end
    
    -- Stop at help text separator
    if line:match('^─+$') then
      break
    end
    
    -- Check if this is a task line
    local task_info = M.parse_task_line(line)
    if task_info.task_id then
      -- Save previous task if exists
      if current_task then
        table.insert(updates, current_task)
      end
      
      current_task = {
        id = task_info.task_id,
        status = task_info.status,
        title = task_info.title,
        description = nil
      }
    elseif current_task then
      -- Check if this is a description line
      local description = M.parse_description_line(line)
      if description and not current_task.description then
        current_task.description = description
      end
    end
    
    ::continue::
  end
  
  -- Save last task
  if current_task then
    table.insert(updates, current_task)
  end
  
  return updates
end

-- Apply updates to tasks
function M.apply_updates(updates)
  local success_count = 0
  local error_count = 0
  
  for _, update in ipairs(updates) do
    if update.id then
      local args = {'edit', update.id}
      
      -- Add updated fields
      if update.title then
        table.insert(args, '--title')
        table.insert(args, update.title)
      end
      
      if update.status then
        table.insert(args, '--status')
        table.insert(args, update.status)
      end
      
      if update.description ~= nil then
        table.insert(args, '--description')
        table.insert(args, update.description or '')
      end
      
      -- Execute update synchronously
      local result = utils.execute_mdtask(args)
      if result then
        success_count = success_count + 1
      else
        error_count = error_count + 1
      end
    end
  end
  
  return success_count, error_count
end

-- Save buffer changes
function M.save_buffer_changes(buf)
  -- Parse buffer for updates
  local updates = M.parse_buffer_for_updates(buf)
  
  if #updates == 0 then
    utils.notify('No tasks found to update', vim.log.levels.WARN)
    return
  end
  
  -- Apply updates
  local success, errors = M.apply_updates(updates)
  
  if errors > 0 then
    utils.notify(string.format('Updated %d tasks, %d errors', success, errors), vim.log.levels.WARN)
  else
    utils.notify(string.format('Successfully updated %d tasks', success))
  end
  
  -- Refresh the task list
  require('mdtask.tasks').list()
end

return M