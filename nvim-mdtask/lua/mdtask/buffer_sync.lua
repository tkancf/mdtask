local M = {}

local utils = require('mdtask.utils')

-- Parse a task line and extract information
function M.parse_task_line(line, line_num)
  -- Get task ID from the line mapping (since it's now virtual text)
  local ui = require('mdtask.ui')
  local task_id = ui.line_to_task_id[line_num]
  
  if not task_id then
    return {}
  end
  
  -- Pattern: - STATUS: Title
  local status, title = line:match('^%s*%- (%w+): (.+)$')
  
  if status and title then
    -- Clean title by trimming
    title = title:match('^%s*(.-)%s*$')
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

-- Parse deadline line
function M.parse_deadline_line(line)
  -- Pattern:     - Deadline: YYYY/MM/DD
  local year, month, day = line:match('^%s+%- Deadline: (%d%d%d%d)/(%d%d)/(%d%d)$')
  if year and month and day then
    -- Convert to mdtask format: YYYY-MM-DD
    return string.format('%s-%s-%s', year, month, day)
  end
  return nil
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
    local task_info = M.parse_task_line(line, i)
    if task_info.task_id then
      -- Save previous task if exists
      if current_task then
        -- If no deadline was set, mark it as removed (empty string)
        if current_task.deadline == nil then
          current_task.deadline = ''
        end
        table.insert(updates, current_task)
      end
      
      current_task = {
        id = task_info.task_id,
        status = task_info.status,
        title = task_info.title,
        description = nil,
        deadline = nil
      }
    elseif current_task then
      -- Check if this is a description line
      local description = M.parse_description_line(line)
      if description and not current_task.description then
        current_task.description = description
      else
        -- Check if this is a deadline line
        local deadline = M.parse_deadline_line(line)
        if deadline then
          current_task.deadline = deadline
        end
      end
    end
    
    ::continue::
  end
  
  -- Save last task
  if current_task then
    -- If no deadline was set, mark it as removed (empty string)
    if current_task.deadline == nil then
      current_task.deadline = ''
    end
    table.insert(updates, current_task)
  end
  
  return updates
end

-- Get original task data by ID
function M.get_original_task_data(task_id)
  local ui = require('mdtask.ui')
  if ui.current_tasks then
    for _, task in ipairs(ui.current_tasks) do
      if task.id == task_id then
        return task
      end
    end
  end
  return nil
end

-- Apply updates to tasks
function M.apply_updates(updates)
  local success_count = 0
  local error_count = 0
  
  for _, update in ipairs(updates) do
    if update.id then
      local args = {'edit', update.id}
      local has_changes = false
      
      -- Get original task to compare
      local original = M.get_original_task_data(update.id)
      
      -- Add updated fields only if they changed
      if update.title and (not original or update.title ~= original.title) then
        table.insert(args, '--title')
        table.insert(args, update.title)
        has_changes = true
      end
      
      if update.status and (not original or update.status ~= original.status) then
        table.insert(args, '--status')
        table.insert(args, update.status)
        has_changes = true
      end
      
      if update.description ~= nil and (not original or update.description ~= (original.description or '')) then
        table.insert(args, '--description')
        table.insert(args, update.description or '')
        has_changes = true
      end
      
      -- Handle deadline changes
      if original and original.deadline then
        -- Original task had a deadline - need to extract just the date part
        local original_date = original.deadline:match('(%d%d%d%d%-%d%d%-%d%d)')
        if update.deadline == '' then
          -- Deadline was removed
          table.insert(args, '--deadline')
          table.insert(args, '')
          has_changes = true
        elseif update.deadline and update.deadline ~= '' and update.deadline ~= original_date then
          -- Deadline was changed
          table.insert(args, '--deadline')
          table.insert(args, update.deadline)
          has_changes = true
        end
      elseif update.deadline and update.deadline ~= '' then
        -- New deadline was added
        table.insert(args, '--deadline')
        table.insert(args, update.deadline)
        has_changes = true
      end
      
      -- Execute update only if there are changes
      if has_changes then
        local result = utils.execute_mdtask(args)
        if result then
          success_count = success_count + 1
        else
          error_count = error_count + 1
        end
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