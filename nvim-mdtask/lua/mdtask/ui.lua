local M = {}

local utils = require('mdtask.utils')
local config = require('mdtask.config')
local highlights = require('mdtask.highlights')
local stats = require('mdtask.stats')

-- View modes
M.view_modes = {
  compact = 'compact',
  detailed = 'detailed'
}

-- Current view mode (default to detailed)
M.current_view_mode = M.view_modes.detailed

-- Sort functions
local sort_functions = {
  default = function(tasks)
    -- Default order (as provided)
    return tasks
  end,
  
  created_asc = function(tasks)
    table.sort(tasks, function(a, b)
      return a.created < b.created
    end)
    return tasks
  end,
  
  created_desc = function(tasks)
    table.sort(tasks, function(a, b)
      return a.created > b.created
    end)
    return tasks
  end,
  
  updated_asc = function(tasks)
    table.sort(tasks, function(a, b)
      return a.updated < b.updated
    end)
    return tasks
  end,
  
  updated_desc = function(tasks)
    table.sort(tasks, function(a, b)
      return a.updated > b.updated
    end)
    return tasks
  end,
  
  title_asc = function(tasks)
    table.sort(tasks, function(a, b)
      return (a.title or ''):lower() < (b.title or ''):lower()
    end)
    return tasks
  end,
  
  title_desc = function(tasks)
    table.sort(tasks, function(a, b)
      return (a.title or ''):lower() > (b.title or ''):lower()
    end)
    return tasks
  end,
  
  status = function(tasks)
    -- Sort by status priority: TODO, WIP, WAIT, SCHE, DONE
    local status_priority = {
      TODO = 1,
      WIP = 2,
      WAIT = 3,
      SCHE = 4,
      DONE = 5,
    }
    
    table.sort(tasks, function(a, b)
      local a_priority = status_priority[a.status] or 99
      local b_priority = status_priority[b.status] or 99
      if a_priority == b_priority then
        -- Secondary sort by updated date
        return a.updated > b.updated
      end
      return a_priority < b_priority
    end)
    return tasks
  end,
  
  deadline = function(tasks)
    -- Tasks with deadline first, then by deadline date
    table.sort(tasks, function(a, b)
      local a_deadline = a.deadline
      local b_deadline = b.deadline
      
      if a_deadline and b_deadline then
        return a_deadline < b_deadline
      elseif a_deadline then
        return true  -- a has deadline, b doesn't
      elseif b_deadline then
        return false  -- b has deadline, a doesn't
      else
        -- Neither has deadline, sort by updated
        return a.updated > b.updated
      end
    end)
    return tasks
  end,
}

M.task_list_buf = nil
M.task_list_win = nil
M.saved_cursor_pos = nil  -- Save cursor position for refresh
M.saved_task_id = nil  -- Save current task ID for cursor restoration
M.current_sort = 'default'  -- Current sort order
M.current_tasks = nil  -- Store current task list for re-sorting
M.line_to_task_id = {}  -- Map line numbers to task IDs

-- Show task list in a floating window
function M.show_task_list(tasks, title)
  title = title or 'mdtask Tasks'
  
  -- Store tasks for re-sorting
  M.current_tasks = tasks
  
  -- Note: Sorting will be applied to root tasks only to maintain hierarchy
  
  -- Add sort indicator to title
  if M.current_sort ~= 'default' then
    local sort_labels = {
      created_asc = 'Created ↑',
      created_desc = 'Created ↓',
      updated_asc = 'Updated ↑',
      updated_desc = 'Updated ↓',
      title_asc = 'Title A-Z',
      title_desc = 'Title Z-A',
      status = 'Status',
      deadline = 'Deadline',
    }
    title = title .. ' [' .. (sort_labels[M.current_sort] or M.current_sort) .. ']'
  end
  
  -- Check if we have an existing valid window and buffer
  local reuse_window = false
  if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) and
     M.task_list_buf and vim.api.nvim_buf_is_valid(M.task_list_buf) then
    reuse_window = true
    -- Save current cursor position and task ID before refresh
    M.saved_cursor_pos = vim.api.nvim_win_get_cursor(M.task_list_win)
    local row = M.saved_cursor_pos[1]
    
    -- Use line mapping to find task ID
    local task_id = M.line_to_task_id[row]
    if task_id then
      M.saved_task_id = task_id
    else
      -- If on link or description line, check previous lines
      for i = row - 1, math.max(1, row - 4), -1 do
        task_id = M.line_to_task_id[i]
        if task_id then
          M.saved_task_id = task_id
          break
        end
      end
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
  
  -- Store info for virtual text
  local deadline_info = {}  -- { line_number = deadline_status }
  local task_id_info = {}  -- { line_number = task_id }
  M.line_to_task_id = {}  -- Reset line to task ID mapping
  
  -- Build a map of parent tasks and their children
  local parent_map = {}  -- parent_id -> list of child tasks
  local root_tasks = {}  -- tasks without parents
  local task_by_id = {}  -- quick lookup map
  
  for _, task in ipairs(tasks) do
    task_by_id[task.id] = task
    if task.parent_id and task.parent_id ~= '' then
      parent_map[task.parent_id] = parent_map[task.parent_id] or {}
      table.insert(parent_map[task.parent_id], task)
    else
      table.insert(root_tasks, task)
    end
  end
  
  -- Function to display task and its children recursively
  local function display_task_tree(task, indent_level)
    indent_level = indent_level or 0
    
    -- Format task with indentation
    local task_lines, deadline_status, task_id = utils.format_task(task, M.current_view_mode)
    
    -- Apply additional indentation for subtasks
    if indent_level > 0 then
      for i, line in ipairs(task_lines) do
        task_lines[i] = string.rep('  ', indent_level) .. line
      end
    end
    
    local main_line_num = #lines + 1  -- Line number for the main task line
    
    -- Store deadline status for virtual text
    if deadline_status then
      deadline_info[main_line_num] = deadline_status
    end
    
    -- Store task ID for virtual text and line mapping
    if task_id and task_id ~= '' then
      task_id_info[main_line_num] = task_id
      M.line_to_task_id[main_line_num] = task_id
    end
    
    for _, line in ipairs(task_lines) do
      table.insert(lines, line)
    end
    
    -- Display children
    if parent_map[task.id] then
      for _, child in ipairs(parent_map[task.id]) do
        display_task_tree(child, indent_level + 1)
      end
    end
  end
  
  -- Apply sorting to root tasks only
  if M.current_sort ~= 'default' and sort_functions[M.current_sort] then
    root_tasks = sort_functions[M.current_sort](root_tasks)
  end
  
  -- Display all root tasks and their subtrees
  for _, task in ipairs(root_tasks) do
    display_task_tree(task, 0)
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
  local mode_indicator = M.current_view_mode == 'compact' and '[Compact]' or '[Detailed]'
  table.insert(lines, 'Keys: <CR> open  s* commands  ? help  q quit  ' .. mode_indicator)
  
  -- Set buffer content
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  -- Make buffer modifiable for direct editing
  vim.api.nvim_buf_set_option(buf, 'modifiable', true)
  
  -- Apply syntax highlights
  highlights.apply_highlights(buf)
  
  -- Apply virtual text
  highlights.apply_deadline_virtual_text(buf, deadline_info)
  highlights.apply_task_id_virtual_text(buf, task_id_info)
  
  -- Only set these options for new buffers
  if not reuse_window then
    -- Remove 'nofile' buftype to allow :w command
    vim.api.nvim_buf_set_option(buf, 'buftype', '')
    vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
    vim.api.nvim_buf_set_option(buf, 'filetype', 'mdtask')
    -- Set buffer name if not already set
    if vim.api.nvim_buf_get_name(buf) == '' then
      vim.api.nvim_buf_set_name(buf, 'mdtask://list')
    end
    
    -- Set up BufWriteCmd autocmd for saving
    vim.api.nvim_create_autocmd('BufWriteCmd', {
      buffer = buf,
      callback = function()
        local buffer_sync = require('mdtask.buffer_sync')
        buffer_sync.save_buffer_changes(buf)
        -- Mark buffer as not modified
        vim.api.nvim_buf_set_option(buf, 'modified', false)
      end,
    })
    
    -- Set up keymaps
    local opts = { buffer = buf, silent = true }
    local actions = require('mdtask.actions')
    
    -- Helper function to get task ID from current or nearby lines
    local function get_task_id_from_position()
      local row = vim.api.nvim_win_get_cursor(0)[1]
      
      -- Check current line first
      local task_id = M.line_to_task_id[row]
      
      -- If not found, check previous lines (for when on link or description line)
      if not task_id then
        -- Check up to 4 lines above (to handle title + link + description)
        for i = row - 1, math.max(1, row - 4), -1 do
          task_id = M.line_to_task_id[i]
          if task_id then break end
        end
      end
      
      return task_id
    end
    
    -- Add 'q' mapping for quick quit
    vim.keymap.set('n', 'q', function()
      vim.api.nvim_win_close(win, true)
    end, opts)
    
    -- <CR> to open the task file
    vim.keymap.set('n', '<CR>', function()
      local task_id = get_task_id_from_position()
      if task_id then
        -- Open the task file in current window
        vim.api.nvim_win_close(win, true)
        -- Get task file path and open it
        local timestamp = task_id:match('task/(.+)')
        if timestamp then
          -- Try different possible paths
          local possible_paths = {
            '_tkancf/' .. timestamp .. '.md',
            timestamp .. '.md',
          }
          
          for _, path in ipairs(possible_paths) do
            if vim.fn.filereadable(path) == 1 then
              vim.cmd('edit ' .. vim.fn.fnameescape(path))
              return
            end
          end
          
          -- If no file found, try the most likely path
          vim.cmd('edit ' .. vim.fn.fnameescape('_tkancf/' .. timestamp .. '.md'))
        end
      end
    end, opts)
  
    -- sa to archive
    vim.keymap.set('n', 'sa', function()
      actions.quick_archive()
    end, opts)
    
    vim.keymap.set('n', 'sr', function()
      require('mdtask.tasks').list()
    end, opts)
    
    -- sn to create new task
    vim.keymap.set('n', 'sn', function()
      -- Don't close the window, just hide it temporarily
      require('mdtask.tasks').new()
    end, opts)
    
    -- se to edit task
    vim.keymap.set('n', 'se', function()
      local task_id = get_task_id_from_position()
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
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'DONE')
      end
    end, opts)
    
    -- st to mark as TODO
    vim.keymap.set('n', 'st', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'TODO')
      end
    end, opts)
    
    -- sw to mark as WIP
    vim.keymap.set('n', 'sw', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'WIP')
      end
    end, opts)
    
    -- sh to mark as SCHE (scheduled)
    vim.keymap.set('n', 'sh', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'SCHE')
      end
    end, opts)
    
    -- sz to mark as WAIT
    vim.keymap.set('n', 'sz', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'WAIT')
      end
    end, opts)
    
    -- Field-specific edits
    -- sS to edit status (with selection dialog)
    vim.keymap.set('n', 'sS', function()
      local task_id = get_task_id_from_position()
      if task_id then
        require('mdtask.tasks').edit_field(task_id, 'status')
      end
    end, opts)
    
    -- sT to edit title
    vim.keymap.set('n', 'sT', function()
      local task_id = get_task_id_from_position()
      if task_id then
        require('mdtask.tasks').edit_field(task_id, 'title')
      end
    end, opts)
    
    -- sD to edit description
    vim.keymap.set('n', 'sD', function()
      local task_id = get_task_id_from_position()
      if task_id then
        require('mdtask.tasks').edit_field(task_id, 'description')
      end
    end, opts)
    
    -- s/ to search tasks
    vim.keymap.set('n', 's/', function()
      vim.ui.input({ prompt = 'Search tasks: ' }, function(query)
        if query and query ~= '' then
          vim.api.nvim_win_close(win, true)
          require('mdtask.tasks').search(query)
        end
      end)
    end, opts)
    
    -- sW to open web interface
    vim.keymap.set('n', 'sW', function()
      require('mdtask.tasks').open_web()
    end, opts)
    
    -- Sort commands
    -- o to show sort menu
    vim.keymap.set('n', 'o', function()
      local sort_options = {
        'default - Default order',
        'created_asc - Created (oldest first)',
        'created_desc - Created (newest first)',
        'updated_asc - Updated (oldest first)', 
        'updated_desc - Updated (newest first)',
        'title_asc - Title (A-Z)',
        'title_desc - Title (Z-A)',
        'status - Status priority',
        'deadline - Deadline (earliest first)',
      }
      
      vim.ui.select(sort_options, {
        prompt = 'Select sort order:',
        format_item = function(item)
          return item
        end,
      }, function(choice)
        if choice then
          local sort_key = choice:match('^(%S+)')
          M.current_sort = sort_key
          -- Re-display with new sort
          if M.current_tasks then
            M.show_task_list(M.current_tasks)
          end
        end
      end)
    end, opts)
    
    -- Quick sort shortcuts
    vim.keymap.set('n', 'oc', function()
      -- Toggle between created asc/desc
      if M.current_sort == 'created_desc' then
        M.current_sort = 'created_asc'
      else
        M.current_sort = 'created_desc'
      end
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    vim.keymap.set('n', 'ou', function()
      -- Toggle between updated asc/desc
      if M.current_sort == 'updated_desc' then
        M.current_sort = 'updated_asc'
      else
        M.current_sort = 'updated_desc'
      end
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    vim.keymap.set('n', 'ot', function()
      -- Toggle between title asc/desc
      if M.current_sort == 'title_asc' then
        M.current_sort = 'title_desc'
      else
        M.current_sort = 'title_asc'
      end
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    vim.keymap.set('n', 'os', function()
      -- Sort by status
      M.current_sort = 'status'
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    vim.keymap.set('n', 'od', function()
      -- Sort by deadline
      M.current_sort = 'deadline'
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    vim.keymap.set('n', 'oO', function()
      -- Reset to default order
      M.current_sort = 'default'
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    -- Task operations
    vim.keymap.set('n', 'sy', function()
      require('mdtask.tasks').copy_task()
    end, opts)
    
    vim.keymap.set('n', 'sP', function()
      require('mdtask.tasks').paste_task()
    end, opts)
    
    vim.keymap.set('n', 'sdd', function()
      require('mdtask.tasks').delete_task()
    end, opts)
    
    -- sN to create subtask
    vim.keymap.set('n', 'sN', function()
      require('mdtask.tasks').new_subtask()
    end, opts)
    
    -- sL to list subtasks
    vim.keymap.set('n', 'sL', function()
      require('mdtask.tasks').list_subtasks()
    end, opts)
    
    -- sv to toggle view mode
    vim.keymap.set('n', 'sv', function()
      -- Toggle between compact and detailed view
      if M.current_view_mode == M.view_modes.compact then
        M.current_view_mode = M.view_modes.detailed
      else
        M.current_view_mode = M.view_modes.compact
      end
      -- Refresh the display
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    -- si to show statistics
    vim.keymap.set('n', 'si', function()
      M.show_stats(M.current_tasks)
    end, opts)
    
    -- ? to show help
    vim.keymap.set('n', '?', function()
      local help_text = [[
Task List Shortcuts:

Navigation & View:
  <CR>    Open task file
  sp      Preview task
  q       Quit
  sr      Refresh list
  s/      Search tasks
  sW      Open web interface
  sv      Toggle view (compact/detailed)
  si      Show task statistics
  ?       Show this help

Task Management:
  sn      New task
  sN      New subtask (for current task)
  se      Edit task (full form)
  sa      Archive task
  sy      Copy task (without ID)
  sP      Paste task (create new)
  sdd     Delete task (with confirmation)
  sL      List subtasks of current task

Quick Status Change:
  ss      Toggle status
  st      Set TODO
  sw      Set WIP (Work in Progress)
  sz      Set WAIT
  sh      Set SCHE (Scheduled)
  sd      Set DONE

Field-Specific Edit:
  sS      Edit status (with dialog)
  sT      Edit title
  sD      Edit description

Sorting:
  o       Sort menu
  oc      Sort by created date (toggle)
  ou      Sort by updated date (toggle)
  ot      Sort by title (toggle)
  os      Sort by status
  od      Sort by deadline
  oO      Reset to default order

Direct Editing:
  :w      Save changes (edit mode)
]]
      -- Create help buffer and window on the right side
      local help_buf = vim.api.nvim_create_buf(false, true)
      vim.api.nvim_buf_set_lines(help_buf, 0, -1, false, vim.split(help_text, '\n'))
      vim.api.nvim_buf_set_option(help_buf, 'modifiable', false)
      vim.api.nvim_buf_set_option(help_buf, 'buftype', 'nofile')
      vim.api.nvim_buf_set_option(help_buf, 'swapfile', false)
      vim.api.nvim_buf_set_option(help_buf, 'bufhidden', 'delete')
      
      -- Save current window
      local current_win = vim.api.nvim_get_current_win()
      
      -- Create vertical split on the right
      vim.cmd('vsplit')
      vim.cmd('wincmd L')  -- Move to the right
      
      -- Set buffer in the new window
      vim.api.nvim_win_set_buf(0, help_buf)
      
      -- Set window width (about 1/3 of screen)
      local width = math.floor(vim.o.columns * 0.35)
      vim.api.nvim_win_set_width(0, width)
      
      -- Set window options
      vim.api.nvim_win_set_option(0, 'number', false)
      vim.api.nvim_win_set_option(0, 'relativenumber', false)
      vim.api.nvim_win_set_option(0, 'cursorline', true)
      vim.api.nvim_win_set_option(0, 'wrap', false)
      
      -- Keymaps for closing help
      local help_opts = { buffer = help_buf, silent = true }
      vim.keymap.set('n', 'q', function()
        vim.api.nvim_win_close(0, true)
        vim.api.nvim_set_current_win(current_win)
      end, help_opts)
      vim.keymap.set('n', '<Esc>', function()
        vim.api.nvim_win_close(0, true)
        vim.api.nvim_set_current_win(current_win)
      end, help_opts)
      vim.keymap.set('n', '?', function()
        vim.api.nvim_win_close(0, true)
        vim.api.nvim_set_current_win(current_win)
      end, help_opts)
    end, opts)
  end  -- end of if not reuse_window
  
  -- Position cursor
  if reuse_window and M.saved_task_id then
    -- Try to find the line with the saved task ID using line mapping
    local found = false
    for line_num, task_id in pairs(M.line_to_task_id) do
      if task_id == M.saved_task_id then
        vim.api.nvim_win_set_cursor(win, {line_num, 0})
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
function M.show_task_form(callback, task, form_title)
  task = task or {}
  
  -- Store form data
  local form_data = {
    title = task.title or '',
    description = task.description or '',
    status = task.status or 'TODO',
    tags = task.tags and table.concat(task.tags, ', ') or '',
    content = task.content or ''
  }
  
  -- Show form title if provided
  if form_title then
    vim.notify(form_title, vim.log.levels.INFO)
  end
  
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
  -- Calculate larger window size (90% of screen)
  local win_width = vim.api.nvim_get_option('columns')
  local win_height = vim.api.nvim_get_option('lines')
  local width = math.floor(win_width * 0.9)
  local height = math.floor(win_height * 0.8)
  
  local buf, win = utils.create_float_win({
    width = width,
    height = height,
  })
  
  local content_lines = {
    '# Task Content',
    '# :w to save, :q to cancel, :wq to save and exit',
    '# ---',
    '',
  }
  
  -- Add existing content or default content for new tasks
  if form_data.content and form_data.content ~= '' then
    for line in form_data.content:gmatch("[^\n]*") do
      table.insert(content_lines, line)
    end
  else
    -- For new tasks, add title as default content
    if form_data.title and form_data.title ~= '' then
      table.insert(content_lines, '# ' .. form_data.title)
      table.insert(content_lines, '')
    end
  end
  
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, content_lines)
  vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
  vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')
  vim.api.nvim_buf_set_option(buf, 'modifiable', true)
  
  -- Set cursor to end of content
  local line_count = vim.api.nvim_buf_line_count(buf)
  vim.api.nvim_win_set_cursor(win, {line_count, 0})
  
  -- Track if content has been saved
  local content_saved = false
  
  -- Create save function
  local function save_content()
    -- Get content (skip header lines)
    local lines = vim.api.nvim_buf_get_lines(buf, 3, -1, false)
    form_data.content = table.concat(lines, '\n'):gsub('^\n+', ''):gsub('\n+$', '')
    
    -- Parse tags into array
    local tags = {}
    if form_data.tags and form_data.tags ~= '' then
      for tag in form_data.tags:gmatch('[^,]+') do
        table.insert(tags, tag:match('^%s*(.-)%s*$'))
      end
    end
    
    -- Mark as saved
    content_saved = true
    vim.api.nvim_buf_set_option(buf, 'modified', false)
    
    -- Store parsed data
    form_data._parsed = {
      title = form_data.title,
      description = form_data.description,
      status = form_data.status,
      tags = tags,
      content = form_data.content
    }
  end
  
  -- Create close function
  local function close_window()
    vim.api.nvim_win_close(win, true)
    
    -- Trigger callback if content was saved
    if content_saved and callback then
      callback(form_data._parsed)
    end
    
    -- Return to task list if it exists
    if M.task_list_win and vim.api.nvim_win_is_valid(M.task_list_win) then
      vim.api.nvim_set_current_win(M.task_list_win)
    end
  end
  
  -- Set buffer name for commands
  vim.api.nvim_buf_set_name(buf, 'mdtask-content')
  
  -- Define commands for this buffer
  vim.api.nvim_buf_create_user_command(buf, 'w', function()
    save_content()
    utils.notify('Content saved')
  end, {})
  
  vim.api.nvim_buf_create_user_command(buf, 'q', function()
    if vim.api.nvim_buf_get_option(buf, 'modified') then
      utils.notify('No write since last change (add ! to override)', vim.log.levels.WARN)
    else
      close_window()
    end
  end, {})
  
  vim.api.nvim_buf_create_user_command(buf, 'q!', function()
    close_window()
  end, { bang = true })
  
  vim.api.nvim_buf_create_user_command(buf, 'wq', function()
    save_content()
    close_window()
  end, {})
  
  vim.api.nvim_buf_create_user_command(buf, 'x', function()
    if vim.api.nvim_buf_get_option(buf, 'modified') then
      save_content()
    end
    close_window()
  end, {})
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

-- Show task statistics
function M.show_stats(tasks)
  local task_stats = stats.calculate_stats(tasks or M.current_tasks or {})
  local lines = stats.format_stats(task_stats)
  
  -- Calculate window size
  local width = 45
  local height = #lines + 2
  
  -- Create floating window
  local buf, win = utils.create_float_win({
    width = width,
    height = height,
  })
  
  -- Set buffer content
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(buf, 'modifiable', false)
  vim.api.nvim_buf_set_option(buf, 'buftype', 'nofile')
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
  
  -- Add highlighting
  local ns_id = vim.api.nvim_create_namespace('mdtask_stats')
  
  -- Highlight headers
  vim.api.nvim_buf_add_highlight(buf, ns_id, 'Title', 0, 0, -1)
  vim.api.nvim_buf_add_highlight(buf, ns_id, 'Comment', 1, 0, -1)
  
  -- Highlight section headers
  for i, line in ipairs(lines) do
    if line:match('^%w+:$') then
      vim.api.nvim_buf_add_highlight(buf, ns_id, 'Type', i-1, 0, -1)
    elseif line:match('^Progress:') then
      vim.api.nvim_buf_add_highlight(buf, ns_id, 'Type', i-1, 0, -1)
    elseif line:match('█') then
      -- Highlight progress bar
      local start_pos = line:find('[')
      local end_pos = line:find(']')
      if start_pos and end_pos then
        vim.api.nvim_buf_add_highlight(buf, ns_id, 'String', i-1, start_pos, end_pos+1)
      end
    end
  end
  
  -- Close on any key
  local opts = { buffer = buf, silent = true }
  vim.keymap.set('n', '<Esc>', function() vim.api.nvim_win_close(win, true) end, opts)
  vim.keymap.set('n', 'q', function() vim.api.nvim_win_close(win, true) end, opts)
  vim.keymap.set('n', '<CR>', function() vim.api.nvim_win_close(win, true) end, opts)
end

return M