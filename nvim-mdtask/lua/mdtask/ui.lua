local M = {}

local utils = require('mdtask.utils')
local config = require('mdtask.config')

M.task_list_buf = nil
M.task_list_win = nil

-- Show task list in a floating window
function M.show_task_list(tasks, title)
  title = title or 'mdtask Tasks'
  
  local buf, win = utils.create_float_win({
    width = 100,
    height = math.min(30, #tasks + 5),
  })
  
  M.task_list_buf = buf
  M.task_list_win = win
  
  -- Prepare lines for display
  local lines = { title, string.rep('─', #title), '' }
  
  for _, task in ipairs(tasks) do
    table.insert(lines, utils.format_task(task))
  end
  
  -- Set buffer content
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(buf, 'modifiable', false)
  vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
  vim.api.nvim_buf_set_option(buf, 'filetype', 'mdtask')
  
  -- Set up keymaps
  local opts = { buffer = buf, silent = true }
  local actions = require('mdtask.actions')
  
  vim.keymap.set('n', 'q', function()
    vim.api.nvim_win_close(win, true)
  end, opts)
  
  vim.keymap.set('n', '<CR>', function()
    local line = vim.api.nvim_get_current_line()
    local task_id = line:match('%(([^)]+)%)')
    if task_id then
      vim.api.nvim_win_close(win, true)
      require('mdtask.tasks').edit(task_id)
    end
  end, opts)
  
  vim.keymap.set('n', 'a', function()
    actions.quick_archive()
  end, opts)
  
  vim.keymap.set('n', 'r', function()
    require('mdtask.tasks').list()
  end, opts)
  
  vim.keymap.set('n', 'n', function()
    vim.api.nvim_win_close(win, true)
    require('mdtask.tasks').new()
  end, opts)
  
  -- New keybindings
  vim.keymap.set('n', 's', function()
    actions.toggle_task_status()
  end, opts)
  
  vim.keymap.set('n', 'p', function()
    actions.preview_task()
  end, opts)
  
  vim.keymap.set('n', 'd', function()
    local line = vim.api.nvim_get_current_line()
    local task_id = line:match('%(([^)]+)%)')
    if task_id then
      actions.quick_status_update(task_id, 'DONE')
    end
  end, opts)
  
  vim.keymap.set('n', 't', function()
    local line = vim.api.nvim_get_current_line()
    local task_id = line:match('%(([^)]+)%)')
    if task_id then
      actions.quick_status_update(task_id, 'TODO')
    end
  end, opts)
  
  vim.keymap.set('n', 'w', function()
    local line = vim.api.nvim_get_current_line()
    local task_id = line:match('%(([^)]+)%)')
    if task_id then
      actions.quick_status_update(task_id, 'WIP')
    end
  end, opts)
  
  -- Position cursor after header
  vim.api.nvim_win_set_cursor(win, {4, 0})
  
  -- Show help
  vim.api.nvim_echo({
    {'Keys: ', 'Normal'},
    {'<CR>', 'Special'}, {' edit, ', 'Normal'},
    {'p', 'Special'}, {' preview, ', 'Normal'},
    {'s', 'Special'}, {' toggle status, ', 'Normal'},
    {'d', 'Special'}, {' done, ', 'Normal'},
    {'t', 'Special'}, {' todo, ', 'Normal'},
    {'w', 'Special'}, {' wip, ', 'Normal'},
    {'a', 'Special'}, {' archive, ', 'Normal'},
    {'n', 'Special'}, {' new, ', 'Normal'},
    {'r', 'Special'}, {' refresh, ', 'Normal'},
    {'q', 'Special'}, {' quit', 'Normal'},
  }, false, {})
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
    if not title then return end  -- User cancelled
    -- Validate title
    title = title:match('^%s*(.-)%s*$') -- trim whitespace
    if title == '' then
      utils.notify('Title is required', vim.log.levels.ERROR)
      return
    end
    form_data.title = title
    
    vim.ui.input({ prompt = 'Description: ', default = form_data.description }, function(description)
      if description == nil then return end  -- User cancelled
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
          if not status then return end  -- User cancelled
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
    '# Press <C-c> or :q to cancel',
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
  end, opts)
  vim.keymap.set('i', '<C-c>', '<Esc>:q<CR>', opts)
  vim.keymap.set('n', 'q', function()
    if not vim.api.nvim_buf_get_option(buf, 'modified') then
      vim.api.nvim_win_close(win, true)
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
  vim.keymap.set('n', 'q', function()
    vim.api.nvim_win_close(win, true)
  end, opts)
  vim.keymap.set('n', '<Esc>', function()
    vim.api.nvim_win_close(win, true)
  end, opts)
end

return M