local M = {}

local utils = require('mdtask.utils')
local ui = require('mdtask.ui')
local config = require('mdtask.config')

-- List all tasks
function M.list()
  local args = {'list'}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to list tasks: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local tasks = utils.parse_json(output)
    if not tasks or #tasks == 0 then
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
  
  local args = {'list', '--status', status}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to list tasks: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local tasks = utils.parse_json(output)
    if not tasks or #tasks == 0 then
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
  
  local args = {'search', query}
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to search tasks: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local tasks = utils.parse_json(output)
    if not tasks or #tasks == 0 then
      utils.notify('No tasks found for: ' .. query)
      return
    end
    
    ui.show_task_list(tasks, 'Search Results: ' .. query)
  end)
end

-- Create new task
function M.new()
  ui.show_task_form(function(task_data)
    local args = {'new', '--format', 'json'}
    
    -- Add title (required)
    if task_data.title and task_data.title ~= '' then
      -- Remove newlines from title
      local cleaned_title = task_data.title:gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
      table.insert(args, '--title')
      table.insert(args, cleaned_title)
    else
      utils.notify('Title is required', vim.log.levels.ERROR)
      return
    end
    
    -- Add description (always provide this flag to avoid interactive prompts)
    -- Remove newlines from description
    local cleaned_description = (task_data.description or ''):gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
    table.insert(args, '--description')
    table.insert(args, cleaned_description)
    
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
      
      -- Parse JSON response to get file path
      local ok, task_response = pcall(vim.json.decode, output)
      if ok and task_response and task_response.file_path then
        utils.notify('Task created successfully')
        
        -- Open the created file in a new tab for editing
        vim.cmd('tabnew ' .. vim.fn.fnameescape(task_response.file_path))
        
        -- Refresh task list if it's open
        if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
          M.list()
        end
      else
        utils.notify('Task created but could not open file for editing', vim.log.levels.WARN)
        -- Refresh task list if it's open
        if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
          M.list()
        end
      end
    end)
  end)
end

-- Edit task
function M.edit(task_id)
  if not task_id or task_id == '' then
    -- Get task ID from current position using line mapping
    local ui = require('mdtask.ui')
    local row = vim.api.nvim_win_get_cursor(0)[1]
    
    -- Check current line and nearby lines
    task_id = ui.line_to_task_id[row]
    if not task_id then
      for i = row - 1, math.max(1, row - 4), -1 do
        task_id = ui.line_to_task_id[i]
        if task_id then break end
      end
    end
    
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
        -- Remove newlines from title
        local cleaned_title = task_data.title:gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
        table.insert(args, '--title')
        table.insert(args, cleaned_title)
      end
      
      if task_data.description ~= nil then
        -- Remove newlines from description
        local cleaned_description = task_data.description:gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
        table.insert(args, '--description')
        table.insert(args, cleaned_description)
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

-- Edit specific field of a task
function M.edit_field(task_id, field, value)
  -- If no task ID, try to get from current position
  if not task_id or task_id == '' then
    local ui = require('mdtask.ui')
    local row = vim.api.nvim_win_get_cursor(0)[1]
    
    -- Check current line and nearby lines
    task_id = ui.line_to_task_id[row]
    if not task_id then
      for i = row - 1, math.max(1, row - 4), -1 do
        task_id = ui.line_to_task_id[i]
        if task_id then break end
      end
    end
    
    if not task_id then
      utils.notify('No task ID found', vim.log.levels.ERROR)
      return
    end
  end
  
  -- If field is provided but no value, prompt for it
  if field and not value then
    if field == 'status' then
      -- Show status selection
      vim.ui.select(
        {'TODO', 'WIP', 'WAIT', 'SCHE', 'DONE'},
        {
          prompt = 'Select status:',
          format_item = function(item)
            return item
          end,
        },
        function(status)
          if status then
            M.edit_field(task_id, 'status', status)
          end
        end
      )
      return
    elseif field == 'title' then
      utils.get_task_by_id(task_id, function(err, task)
        if err then
          utils.notify('Failed to get task: ' .. err, vim.log.levels.ERROR)
          return
        end
        
        vim.ui.input({ 
          prompt = 'Title: ', 
          default = task.title 
        }, function(title)
          if title then
            M.edit_field(task_id, 'title', title)
          end
        end)
      end)
      return
    elseif field == 'description' then
      utils.get_task_by_id(task_id, function(err, task)
        if err then
          utils.notify('Failed to get task: ' .. err, vim.log.levels.ERROR)
          return
        end
        
        vim.ui.input({ 
          prompt = 'Description: ', 
          default = task.description or ''
        }, function(description)
          if description ~= nil then
            M.edit_field(task_id, 'description', description)
          end
        end)
      end)
      return
    end
  end
  
  -- If both field and value are provided, update directly
  if field and value then
    local args = {'edit', task_id}
    
    if field == 'status' then
      table.insert(args, '--status')
      table.insert(args, value)
    elseif field == 'title' then
      local cleaned_title = value:gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
      table.insert(args, '--title')
      table.insert(args, cleaned_title)
    elseif field == 'description' then
      local cleaned_description = value:gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
      table.insert(args, '--description')
      table.insert(args, cleaned_description)
    else
      utils.notify('Invalid field: ' .. field, vim.log.levels.ERROR)
      return
    end
    
    utils.execute_mdtask(args, function(err, output)
      if err then
        utils.notify('Failed to update task: ' .. err, vim.log.levels.ERROR)
        return
      end
      
      utils.notify('Task ' .. field .. ' updated successfully')
      
      -- Refresh task list if it's open
      if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
        M.list()
      end
    end)
  else
    -- No field specified, show full edit form
    M.edit(task_id)
  end
end

-- Archive task
function M.archive(task_id)
  if not task_id or task_id == '' then
    -- Get task ID from current position using line mapping
    local ui = require('mdtask.ui')
    local row = vim.api.nvim_win_get_cursor(0)[1]
    
    -- Check current line and nearby lines
    task_id = ui.line_to_task_id[row]
    if not task_id then
      for i = row - 1, math.max(1, row - 4), -1 do
        task_id = ui.line_to_task_id[i]
        if task_id then break end
      end
    end
    
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

-- Copy task (store in module-level clipboard)
M.task_clipboard = nil

-- Copy task
function M.copy_task(task_id)
  if not task_id or task_id == '' then
    -- Get task ID from current position using line mapping
    local ui = require('mdtask.ui')
    local row = vim.api.nvim_win_get_cursor(0)[1]
    
    -- Check current line and nearby lines
    task_id = ui.line_to_task_id[row]
    if not task_id then
      for i = row - 1, math.max(1, row - 4), -1 do
        task_id = ui.line_to_task_id[i]
        if task_id then break end
      end
    end
    
    if not task_id then
      utils.notify('No task found to copy', vim.log.levels.ERROR)
      return
    end
  end
  
  utils.get_task_by_id(task_id, function(err, task)
    if err then
      utils.notify('Failed to get task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    -- Store task data without ID
    M.task_clipboard = {
      title = task.title,
      description = task.description,
      status = task.status,
      tags = vim.deepcopy(task.tags),
      content = task.content,
      deadline = task.deadline
    }
    
    utils.notify('Task copied')
  end)
end

-- Paste task
function M.paste_task()
  if not M.task_clipboard then
    utils.notify('No task in clipboard', vim.log.levels.ERROR)
    return
  end
  
  local args = {'new'}
  
  -- Add title (required)
  if M.task_clipboard.title then
    table.insert(args, '--title')
    table.insert(args, M.task_clipboard.title)
  end
  
  -- Add description
  table.insert(args, '--description')
  table.insert(args, M.task_clipboard.description or '')
  
  -- Add content
  table.insert(args, '--content')
  table.insert(args, M.task_clipboard.content or '')
  
  -- Add status
  if M.task_clipboard.status then
    table.insert(args, '--status')
    table.insert(args, M.task_clipboard.status)
  end
  
  -- Add tags
  if M.task_clipboard.tags and #M.task_clipboard.tags > 0 then
    -- Filter out mdtask system tags
    local user_tags = {}
    for _, tag in ipairs(M.task_clipboard.tags) do
      if not tag:match('^mdtask/') then
        table.insert(user_tags, tag)
      end
    end
    if #user_tags > 0 then
      table.insert(args, '--tags')
      table.insert(args, table.concat(user_tags, ','))
    end
  end
  
  -- Add deadline if present
  if M.task_clipboard.deadline then
    -- Extract just the date part from deadline
    local deadline_date = M.task_clipboard.deadline:match('(%d%d%d%d%-%d%d%-%d%d)')
    if deadline_date then
      table.insert(args, '--deadline')
      table.insert(args, deadline_date)
    end
  end
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to paste task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    utils.notify('Task pasted successfully')
    
    -- Refresh task list if it's open
    if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
      M.list()
    end
  end)
end

-- Delete task
function M.delete_task(task_id)
  if not task_id or task_id == '' then
    -- Get task ID from current position using line mapping
    local ui = require('mdtask.ui')
    local row = vim.api.nvim_win_get_cursor(0)[1]
    
    -- Check current line and nearby lines
    task_id = ui.line_to_task_id[row]
    if not task_id then
      for i = row - 1, math.max(1, row - 4), -1 do
        task_id = ui.line_to_task_id[i]
        if task_id then break end
      end
    end
    
    if not task_id then
      utils.notify('No task found to delete', vim.log.levels.ERROR)
      return
    end
  end
  
  -- Get task details for confirmation
  utils.get_task_by_id(task_id, function(err, task)
    if err then
      utils.notify('Failed to get task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    -- Confirm deletion
    vim.ui.select(
      {'Yes', 'No'},
      {
        prompt = string.format('Delete task "%s"?', task.title),
        format_item = function(item)
          return item
        end,
      },
      function(choice)
        if choice == 'Yes' then
          local args = {'delete', task_id}
          
          utils.execute_mdtask(args, function(err, output)
            if err then
              utils.notify('Failed to delete task: ' .. err, vim.log.levels.ERROR)
              return
            end
            
            utils.notify('Task deleted successfully')
            
            -- Refresh task list if it's open
            if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
              M.list()
            end
          end)
        end
      end
    )
  end)
end

-- Create subtask
function M.new_subtask(parent_id)
  -- Save current buffer and cursor position before any operations
  local original_buf = vim.api.nvim_get_current_buf()
  local original_win = vim.api.nvim_get_current_win()
  local original_cursor = vim.api.nvim_win_get_cursor(original_win)
  
  if not parent_id or parent_id == '' then
    -- First try to get parent task ID from current position (task list buffer)
    local ui = require('mdtask.ui')
    local row = vim.api.nvim_win_get_cursor(0)[1]
    
    parent_id = ui.line_to_task_id[row]
    if not parent_id then
      for i = row - 1, math.max(1, row - 4), -1 do
        parent_id = ui.line_to_task_id[i]
        if parent_id then break end
      end
    end
    
    -- If not found in task list, try to get from current buffer file path
    if not parent_id then
      parent_id = utils.get_task_id_from_buffer()
    end
    
    if not parent_id then
      utils.notify('No parent task found. Open a task file or use this command from the task list.', vim.log.levels.ERROR)
      return
    end
  end
  
  -- Get parent task details first
  utils.get_task_by_id(parent_id, function(err, parent_task)
    if err then
      utils.notify('Failed to get parent task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    -- Show task form for subtask
    ui.show_task_form(function(task_data)
      local args = {'new', '--parent', parent_id, '--format', 'json'}
      
      -- Add title (required)
      if task_data.title and task_data.title ~= '' then
        local cleaned_title = task_data.title:gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
        table.insert(args, '--title')
        table.insert(args, cleaned_title)
      else
        utils.notify('Title is required', vim.log.levels.ERROR)
        return
      end
      
      -- Add description
      local cleaned_description = (task_data.description or ''):gsub('[\n\r]+', ' '):gsub('%s+', ' '):match('^%s*(.-)%s*$')
      table.insert(args, '--description')
      table.insert(args, cleaned_description)
      
      -- Add content
      table.insert(args, '--content')
      table.insert(args, task_data.content or '')
      
      -- Add status (inherit from parent if not specified)
      if task_data.status and task_data.status ~= '' then
        table.insert(args, '--status')
        table.insert(args, task_data.status)
      else
        table.insert(args, '--status')
        table.insert(args, parent_task.status or 'TODO')
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
          utils.notify('Failed to create subtask: ' .. err, vim.log.levels.ERROR)
          return
        end
        
        -- Parse JSON response
        local ok, task_data = pcall(vim.json.decode, output)
        if not ok or not task_data then
          utils.notify('Failed to parse subtask response', vim.log.levels.ERROR)
          return
        end
        
        utils.notify(string.format('Subtask created for "%s"', parent_task.title))
        
        -- Insert markdown link at original cursor position
        if task_data.title and task_data.file_path then
          -- Switch back to original buffer and window
          vim.api.nvim_set_current_win(original_win)
          vim.api.nvim_set_current_buf(original_buf)
          
          -- Create markdown link
          local link = string.format('[%s](%s)', task_data.title, task_data.file_path)
          
          -- Insert at cursor position
          local row = original_cursor[1]
          local col = original_cursor[2]
          local lines = vim.api.nvim_buf_get_lines(original_buf, row - 1, row, false)
          local line = lines[1] or ''
          
          -- Insert the link at cursor position
          local new_line = line:sub(1, col) .. link .. line:sub(col + 1)
          vim.api.nvim_buf_set_lines(original_buf, row - 1, row, false, {new_line})
          
          -- Move cursor after the inserted link
          vim.api.nvim_win_set_cursor(original_win, {row, col + #link})
        end
        
        -- Refresh task list if it's open
        if ui.task_list_buf and vim.api.nvim_buf_is_valid(ui.task_list_buf) then
          M.list()
        end
      end)
    end, nil, string.format('New Subtask for: %s', parent_task.title))
  end)
end

-- List subtasks of a parent task
function M.list_subtasks(parent_id)
  if not parent_id or parent_id == '' then
    -- Get parent task ID from current position
    local ui = require('mdtask.ui')
    local row = vim.api.nvim_win_get_cursor(0)[1]
    
    parent_id = ui.line_to_task_id[row]
    if not parent_id then
      for i = row - 1, math.max(1, row - 4), -1 do
        parent_id = ui.line_to_task_id[i]
        if parent_id then break end
      end
    end
    
    if not parent_id then
      utils.notify('No parent task found', vim.log.levels.ERROR)
      return
    end
  end
  
  -- Get parent task details
  utils.get_task_by_id(parent_id, function(err, parent_task)
    if err then
      utils.notify('Failed to get parent task: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local args = {'list', '--parent', parent_id}
    
    utils.execute_mdtask(args, function(err, output)
      if err then
        utils.notify('Failed to list subtasks: ' .. err, vim.log.levels.ERROR)
        return
      end
      
      local tasks = utils.parse_json(output)
      if not tasks or #tasks == 0 then
        utils.notify(string.format('No subtasks found for "%s"', parent_task.title))
        return
      end
      
      ui.show_task_list(tasks, string.format('Subtasks of: %s', parent_task.title))
    end)
  end)
end

return M