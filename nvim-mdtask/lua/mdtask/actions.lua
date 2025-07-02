local M = {}

local utils = require('mdtask.utils')
local ui = require('mdtask.ui')

-- Quick status update
function M.quick_status_update(task_id, new_status)
  if not task_id or task_id == '' then
    utils.notify('Task ID is required', vim.log.levels.ERROR)
    return
  end
  
  local valid_statuses = {'TODO', 'WIP', 'WAIT', 'SCHE', 'DONE'}
  if not vim.tbl_contains(valid_statuses, new_status) then
    utils.notify('Invalid status: ' .. new_status, vim.log.levels.ERROR)
    return
  end
  
  local args = {'edit', task_id, '--status', new_status}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to update task status: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    utils.notify('Task status updated to ' .. new_status)
    
    -- Refresh task list if it's open
    if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
      require('mdtask.tasks').list()
    end
  end, nil, true)  -- Skip JSON format for edit command
end

-- Toggle task status
function M.toggle_task_status(task_id)
  if not task_id or task_id == '' then
    -- Try to get task ID from current line
    local current_line = vim.api.nvim_get_current_line()
    task_id = current_line:match('%(([^)]+)%)')
    
    if not task_id then
      utils.notify('No task ID found on current line', vim.log.levels.ERROR)
      return
    end
  end
  
  -- Get current task to determine next status
  utils.get_task_by_id(task_id, function(err, task)
    if err then
      utils.notify('Failed to get task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    -- Define status progression
    local status_flow = {
      TODO = 'WIP',
      WIP = 'DONE',
      DONE = 'TODO',
      WAIT = 'WIP',
      SCHE = 'WIP'
    }
    
    local current_status = task.status
    local new_status = status_flow[current_status] or 'TODO'
    
    M.quick_status_update(task_id, new_status)
  end)
end

-- Archive task quickly
function M.quick_archive(task_id)
  if not task_id or task_id == '' then
    -- Try to get task ID from current line
    local current_line = vim.api.nvim_get_current_line()
    task_id = current_line:match('%(([^)]+)%)')
    
    if not task_id then
      utils.notify('No task ID found on current line', vim.log.levels.ERROR)
      return
    end
  end
  
  local args = {'archive', task_id}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to archive task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    utils.notify('Task archived')
    
    -- Refresh task list if it's open
    if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
      require('mdtask.tasks').list()
    end
  end, nil, true)  -- Skip JSON format for archive command
end

-- Show task details in preview window
function M.preview_task(task_id)
  if not task_id or task_id == '' then
    -- Try to get task ID from current line
    local current_line = vim.api.nvim_get_current_line()
    task_id = current_line:match('%(([^)]+)%)')
    
    if not task_id then
      return
    end
  end
  
  utils.get_task_by_id(task_id, function(err, task)
    if err then
      return
    end
    
    ui.show_task_preview(task)
  end)
end

return M