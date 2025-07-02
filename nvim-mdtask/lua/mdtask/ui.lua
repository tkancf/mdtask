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
    local line = vim.api.nvim_get_current_line()
    local task_id = line:match('%(([^)]+)%)')
    if task_id then
      require('mdtask.tasks').archive(task_id)
    end
  end, opts)
  
  vim.keymap.set('n', 'r', function()
    require('mdtask.tasks').list()
  end, opts)
  
  vim.keymap.set('n', 'n', function()
    vim.api.nvim_win_close(win, true)
    require('mdtask.tasks').new()
  end, opts)
  
  -- Position cursor after header
  vim.api.nvim_win_set_cursor(win, {4, 0})
  
  -- Show help
  vim.api.nvim_echo({
    {'Keys: ', 'Normal'},
    {'<CR>', 'Special'}, {' edit, ', 'Normal'},
    {'a', 'Special'}, {' archive, ', 'Normal'},
    {'n', 'Special'}, {' new, ', 'Normal'},
    {'r', 'Special'}, {' refresh, ', 'Normal'},
    {'q', 'Special'}, {' quit', 'Normal'},
  }, false, {})
end

-- Show task creation/editing form
function M.show_task_form(callback, task)
  task = task or {}
  
  local buf, win = utils.create_float_win({
    width = 80,
    height = 20,
  })
  
  local form_lines = {
    '# Task Form',
    '',
    'Title: ' .. (task.title or ''),
    '',
    'Description: ' .. (task.description or ''),
    '',
    'Status: ' .. (task.status or 'TODO'),
    '',
    'Tags: ' .. (task.tags and table.concat(task.tags, ', ') or ''),
    '',
    '# Content',
    '',
    (task.content or ''),
    '',
    '# Instructions',
    '- Edit the fields above',
    '- Press <C-s> to save',
    '- Press <C-c> to cancel',
  }
  
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, form_lines)
  vim.api.nvim_buf_set_option(buf, 'buftype', 'acwrite')
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
  vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')
  
  -- Position cursor on title line
  vim.api.nvim_win_set_cursor(win, {3, #'Title: '})
  
  -- Set up keymaps
  local opts = { buffer = buf, silent = true }
  
  vim.keymap.set('n', '<C-s>', function()
    M.save_task_form(buf, win, callback)
  end, opts)
  
  vim.keymap.set('i', '<C-s>', function()
    M.save_task_form(buf, win, callback)
  end, opts)
  
  vim.keymap.set('n', '<C-c>', function()
    vim.api.nvim_win_close(win, true)
  end, opts)
  
  vim.keymap.set('i', '<C-c>', function()
    vim.api.nvim_win_close(win, true)
  end, opts)
end

-- Save task form data
function M.save_task_form(buf, win, callback)
  local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
  
  local task_data = {}
  local content_lines = {}
  local in_content = false
  
  for _, line in ipairs(lines) do
    local title_match = line:match('^Title: (.*)')
    if title_match then
      task_data.title = title_match:match('^%s*(.-)%s*$')  -- trim whitespace
    end
    
    local desc_match = line:match('^Description: (.*)')
    if desc_match then
      task_data.description = desc_match:match('^%s*(.-)%s*$')
    end
    
    local status_match = line:match('^Status: (.*)')
    if status_match then
      task_data.status = status_match:match('^%s*(.-)%s*$')
    end
    
    local tags_match = line:match('^Tags: (.*)')
    if tags_match then
      local tags_str = tags_match:match('^%s*(.-)%s*$')
      if tags_str and tags_str ~= '' then
        task_data.tags = vim.split(tags_str, ',')
        -- Trim whitespace from each tag
        for i, tag in ipairs(task_data.tags) do
          task_data.tags[i] = tag:match('^%s*(.-)%s*$')
        end
      end
    end
    
    if line == '# Content' then
      in_content = true
    elseif line == '# Instructions' then
      in_content = false
    elseif in_content and line ~= '' then
      table.insert(content_lines, line)
    end
  end
  
  if #content_lines > 0 then
    task_data.content = table.concat(content_lines, '\n')
  end
  
  vim.api.nvim_win_close(win, true)
  
  if callback then
    callback(task_data)
  end
end

return M