local M = {}

local utils = require('mdtask.utils')
local config = require('mdtask.config')

M.task_list_buf = nil
M.task_list_win = nil
M.saved_cursor_pos = nil  -- Save cursor position for refresh
M.saved_task_id = nil  -- Save current task ID for cursor restoration

-- Show task list in a floating window
function M.show_task_list(tasks, title)
  title = title or 'mdtask Tasks'
  
  -- Check if we have an existing valid window and buffer
  local reuse_window = false
  if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) and
     M.task_list_buf and vim.api.nvim_buf_is_valid(M.task_list_buf) then
    reuse_window = true
    -- Save current cursor position and task ID before refresh
    M.saved_cursor_pos = vim.api.nvim_win_get_cursor(M.task_list_win)
    local current_line = vim.api.nvim_get_current_line()
    local task_id = current_line:match('%(([^)]+)%)')
    if task_id then
      M.saved_task_id = task_id
    end
  end
  
  local buf, win
  if reuse_window then
    -- Reuse existing window and buffer
    buf = M.task_list_buf
    win = M.task_list_win
    -- Make buffer modifiable for updating
    vim.api.nvim_buf_set_option(buf, 'modifiable', true)
  else
    -- Calculate window size - almost full screen with some padding
    local win_width = vim.api.nvim_get_option('columns')
    local win_height = vim.api.nvim_get_option('lines')
    
    local width = math.floor(win_width * 0.9)  -- 90% of screen width
    local height = math.floor(win_height * 0.85)  -- 85% of screen height
    
    buf, win = utils.create_float_win({
      width = width,
      height = height,
    })
    
    M.task_list_buf = buf
    M.task_list_win = win
  end
  
  -- Prepare lines for display
  local lines = { title, string.rep('─', #title), '' }
  
  for _, task in ipairs(tasks) do
    table.insert(lines, utils.format_task(task))
  end
  
  -- Add empty lines to fill the window if needed
  local content_lines = #lines
  -- Get window height dynamically for reused windows
  local win_height = vim.api.nvim_win_get_height(win)
  local available_height = win_height - 4  -- Reserve space for help text
  while #lines < available_height do
    table.insert(lines, '')
  end
  
  -- Add help text at the bottom
  local win_width = vim.api.nvim_win_get_width(win)
  table.insert(lines, string.rep('─', math.min(win_width - 2, 80)))
  table.insert(lines, 'Keys: <CR> open  sp preview  ss toggle  st todo  sw wip  sd done  sa archive  sn new  se edit  r refresh  q quit')
  
  -- Set buffer content
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(buf, 'modifiable', false)
  
  -- Only set these options for new buffers
  if not reuse_window then
    vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
    vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
    vim.api.nvim_buf_set_option(buf, 'filetype', 'mdtask')
    
    -- Set up keymaps
    local opts = { buffer = buf, silent = true }
    local actions = require('mdtask.actions')
    
    -- Add 'q' mapping for quick quit
    vim.keymap.set('n', 'q', function()
      vim.api.nvim_win_close(win, true)
    end, opts)
    
    -- <CR> to open the task file
    vim.keymap.set('n', '<CR>', function()
      local line = vim.api.nvim_get_current_line()
      local task_id = line:match('%(([^)]+)%)')
      if task_id then
        -- Open the task file in current window
        vim.api.nvim_win_close(win, true)
        -- Get task file path and open it
        local timestamp = task_id:match('task/(.+)')
        if timestamp then
          vim.cmd('edit ' .. timestamp .. '.md')
        end
      end
    end, opts)
  
    -- sa to archive
    vim.keymap.set('n', 'sa', function()
      actions.quick_archive()
    end, opts)
    
    vim.keymap.set('n', 'r', function()
      require('mdtask.tasks').list()
    end, opts)
    
    -- sn to create new task
    vim.keymap.set('n', 'sn', function()
      -- Don't close the window, just hide it temporarily
      require('mdtask.tasks').new()
    end, opts)
    
    -- se to edit task
    vim.keymap.set('n', 'se', function()
      local line = vim.api.nvim_get_current_line()
      local task_id = line:match('%(([^)]+)%)')
      if task_id then
        require('mdtask.tasks').edit(task_id)
      end
    end, opts)
    
    -- ss to toggle status
    vim.keymap.set('n', 'ss', function()
      actions.toggle_task_status()
    end, opts)
    
    -- sp to preview
    vim.keymap.set('n', 'sp', function()
      actions.preview_task()
    end, opts)
    
    -- sd to mark as DONE
    vim.keymap.set('n', 'sd', function()
      local line = vim.api.nvim_get_current_line()
      local task_id = line:match('%(([^)]+)%)')
      if task_id then
        actions.quick_status_update(task_id, 'DONE')
      end
    end, opts)
    
    -- st to mark as TODO
    vim.keymap.set('n', 'st', function()
      local line = vim.api.nvim_get_current_line()
      local task_id = line:match('%(([^)]+)%)')
      if task_id then
        actions.quick_status_update(task_id, 'TODO')
      end
    end, opts)
    
    -- sw to mark as WIP
    vim.keymap.set('n', 'sw', function()
      local line = vim.api.nvim_get_current_line()
      local task_id = line:match('%(([^)]+)%)')
      if task_id then
        actions.quick_status_update(task_id, 'WIP')
      end
    end, opts)
  end  -- end of if not reuse_window
  
  -- Position cursor
  if reuse_window and M.saved_task_id then
    -- Try to find the line with the saved task ID
    local found = false
    local lines_content = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
    for i, line in ipairs(lines_content) do
      if line:match('%((' .. vim.pesc(M.saved_task_id) .. ')%)') then
        vim.api.nvim_win_set_cursor(win, {i, 0})
        found = true
        break
      end
    end
    
    -- If task not found, try to restore by line number
    if not found and M.saved_cursor_pos then
      local max_line = vim.api.nvim_buf_line_count(buf)
      local row = math.min(M.saved_cursor_pos[1], max_line)
      vim.api.nvim_win_set_cursor(win, {row, 0})
    end
    
    -- Clear saved position and task ID
    M.saved_cursor_pos = nil
    M.saved_task_id = nil
  else
    -- Default position after header
    vim.api.nvim_win_set_cursor(win, {4, 0})
  end
end

-- Show task creation/editing form
function M.show_task_form(callback, task)
  task = task or {}
  
  -- Store form data
  local form_data = {
    title = task.title or '',
    description = task.description or '',
    status = task.status or 'TODO',
    tags = task.tags and table.concat(task.tags, ', ') or '',
    content = task.content or ''
  }
  
  -- Show input prompts sequentially
  vim.ui.input({ prompt = 'Title: ', default = form_data.title }, function(title)
    if not title then 
      -- User cancelled - return to task list
      if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) then
        vim.api.nvim_set_current_win(M.task_list_win)
      end
      return 
    end
    -- Validate title
    title = title:match('^%s*(.-)%s*$') -- trim whitespace
    if title == '' then
      utils.notify('Title is required', vim.log.levels.ERROR)
      return
    end
    form_data.title = title
    
    vim.ui.input({ prompt = 'Description: ', default = form_data.description }, function(description)
      if description == nil then 
        -- User cancelled - return to task list
        if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) then
          vim.api.nvim_set_current_win(M.task_list_win)
        end
        return 
      end
      form_data.description = description
      
      vim.ui.select(
        {'TODO', 'WIP', 'WAIT', 'SCHE', 'DONE'},
        {
          prompt = 'Status:',
          format_item = function(item)
            return item
          end,
        },
        function(status)
          if not status then 
            -- User cancelled - return to task list
            if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) then
              vim.api.nvim_set_current_win(M.task_list_win)
            end
            return 
          end
          form_data.status = status
          
          -- Skip tags input and go directly to content editor
          form_data.tags = ''
          
          -- Show content editor in a buffer
          M.show_content_editor(form_data, callback)
        end
      )
    end)
  end)
end

-- Show content editor in a floating window
function M.show_content_editor(form_data, callback)
  local buf, win = utils.create_float_win({
    width = 80,
    height = 20,
  })
  
  local content_lines = {
    '# Task Content',
    '# Press <C-s> to save and create task',
    '# Press <C-c>, <Esc> or :q to cancel',
    '# ---',
    '',
  }
  
  -- Add existing content
  if form_data.content and form_data.content ~= '' then
    for line in form_data.content:gmatch("[^\n]*") do
      table.insert(content_lines, line)
    end
  end
  
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, content_lines)
  vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
  vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')
  vim.api.nvim_buf_set_option(buf, 'modifiable', true)
  
  -- Set cursor to first content line
  vim.api.nvim_win_set_cursor(win, {5, 0})
  
  -- Create save function
  local function save_and_close()
    -- Get content (skip header lines)
    local lines = vim.api.nvim_buf_get_lines(buf, 4, -1, false)
    form_data.content = table.concat(lines, '\n'):gsub('^\n+', ''):gsub('\n+$', '')
    
    -- Parse tags into array
    local tags = {}
    if form_data.tags and form_data.tags ~= '' then
      for tag in form_data.tags:gmatch('[^,]+') do
        table.insert(tags, tag:match('^%s*(.-)%s*$'))
      end
    end
    
    -- Close window
    vim.api.nvim_win_close(win, true)
    
    -- Trigger callback with parsed data
    if callback then
      callback({
        title = form_data.title,
        description = form_data.description,
        status = form_data.status,
        tags = tags,
        content = form_data.content
      })
    end
  end
  
  -- Set up keymaps
  local opts = { buffer = buf, silent = true }
  
  -- Save shortcuts
  vim.keymap.set('n', '<C-s>', save_and_close, opts)
  vim.keymap.set('i', '<C-s>', save_and_close, opts)
  
  -- Cancel shortcuts
  vim.keymap.set('n', '<C-c>', function()
    vim.api.nvim_win_close(win, true)
    -- Return to task list if it exists
    if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) then
      vim.api.nvim_set_current_win(M.task_list_win)
    end
  end, opts)
  vim.keymap.set('i', '<C-c>', '<Esc>:q<CR>', opts)
  
  -- Add Esc key to cancel and return to task list
  vim.keymap.set('n', '<Esc>', function()
    vim.api.nvim_win_close(win, true)
    -- Return to task list if it exists
    if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) then
      vim.api.nvim_set_current_win(M.task_list_win)
    end
  end, opts)
  
  -- Add 'q' mapping for quick quit (only in normal mode, when not modified)
  vim.keymap.set('n', 'q', function()
    if not vim.api.nvim_buf_get_option(buf, 'modified') then
      vim.api.nvim_win_close(win, true)
      -- Return to task list if it exists
      if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) then
        vim.api.nvim_set_current_win(M.task_list_win)
      end
    end
  end, opts)
end


-- Show task preview in floating window
function M.show_task_preview(task)
  local lines = {
    '# ' .. task.title,
    '',
    '**Status:** ' .. task.status,
    '**ID:** ' .. task.id,
    '**Created:** ' .. task.created:sub(1, 10),
    '**Updated:** ' .. task.updated:sub(1, 10),
  }
  
  if task.description and task.description ~= '' then
    table.insert(lines, '')
    table.insert(lines, '**Description:** ' .. task.description)
  end
  
  if task.deadline then
    table.insert(lines, '**Deadline:** ' .. task.deadline:sub(1, 10))
  end
  
  if task.reminder then
    table.insert(lines, '**Reminder:** ' .. task.reminder:sub(1, 16))
  end
  
  if task.tags and #task.tags > 0 then
    table.insert(lines, '')
    table.insert(lines, '**Tags:** ' .. table.concat(task.tags, ', '))
  end
  
  if task.content and task.content ~= '' then
    table.insert(lines, '')
    table.insert(lines, '---')
    table.insert(lines, '')
    for line in task.content:gmatch("[^\n]*") do
      table.insert(lines, line)
    end
  end
  
  local buf, win = utils.create_float_win({
    width = math.min(80, math.floor(vim.o.columns * 0.8)),
    height = math.min(#lines + 2, math.floor(vim.o.lines * 0.8)),
  })
  
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(buf, 'modifiable', false)
  vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
  vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')
  
  -- Set up keymaps
  local opts = { buffer = buf, silent = true }
  -- Add 'q' mapping for quick quit
  vim.keymap.set('n', 'q', function()
    vim.api.nvim_win_close(win, true)
  end, opts)
  vim.keymap.set('n', '<Esc>', function()
    vim.api.nvim_win_close(win, true)
  end, opts)
end

return M