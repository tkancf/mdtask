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

-- Filter settings
M.current_filters = {
  status = nil,  -- nil = all, or specific status
  tags = {},     -- empty = all, or list of tags to include
  deadline = nil, -- nil = all, 'overdue', 'today', 'week', 'none'
  search = nil   -- nil = no search, or search term
}

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

-- Filter functions
local function apply_filters(tasks)
  local filtered = {}
  local today = os.date('%Y-%m-%d')
  
  for _, task in ipairs(tasks) do
    local include = true
    
    -- Status filter
    if M.current_filters.status and task.status ~= M.current_filters.status then
      include = false
    end
    
    -- Tags filter
    if #M.current_filters.tags > 0 then
      local has_required_tag = false
      for _, required_tag in ipairs(M.current_filters.tags) do
        for _, task_tag in ipairs(task.tags or {}) do
          if task_tag == required_tag then
            has_required_tag = true
            break
          end
        end
        if has_required_tag then break end
      end
      if not has_required_tag then
        include = false
      end
    end
    
    -- Deadline filter
    if M.current_filters.deadline then
      if M.current_filters.deadline == 'none' then
        if task.deadline then include = false end
      elseif M.current_filters.deadline == 'overdue' then
        if not task.deadline or task.deadline >= today then include = false end
      elseif M.current_filters.deadline == 'today' then
        if not task.deadline or task.deadline ~= today then include = false end
      elseif M.current_filters.deadline == 'week' then
        if not task.deadline then include = false
        else
          local deadline_time = os.time{year=task.deadline:sub(1,4), month=task.deadline:sub(6,7), day=task.deadline:sub(9,10)}
          local today_time = os.time()
          local week_time = today_time + (7 * 24 * 60 * 60)
          if deadline_time > week_time then include = false end
        end
      end
    end
    
    -- Search filter
    if M.current_filters.search and M.current_filters.search ~= '' then
      local search_term = M.current_filters.search:lower()
      local found = false
      
      -- Search in title
      if task.title and task.title:lower():find(search_term, 1, true) then
        found = true
      end
      
      -- Search in description
      if not found and task.description and task.description:lower():find(search_term, 1, true) then
        found = true
      end
      
      -- Search in content
      if not found and task.content and task.content:lower():find(search_term, 1, true) then
        found = true
      end
      
      -- Search in tags
      if not found and task.tags then
        for _, tag in ipairs(task.tags) do
          if tag:lower():find(search_term, 1, true) then
            found = true
            break
          end
        end
      end
      
      if not found then
        include = false
      end
    end
    
    if include then
      table.insert(filtered, task)
    end
  end
  
  return filtered
end

M.task_list_buf = nil
M.task_list_win = nil
M.help_win = nil  -- Track help window to close it when main window closes
M.saved_cursor_pos = nil  -- Save cursor position for refresh
M.saved_task_id = nil  -- Save current task ID for cursor restoration
M.current_sort = 'default'  -- Current sort order
M.current_tasks = nil  -- Store current task list for re-sorting
M.line_to_task_id = {}  -- Map line numbers to task IDs

-- Show task list in a floating window
function M.show_task_list(tasks, title)
  title = title or 'mdtask Tasks'
  
  -- Store original tasks for re-filtering/sorting
  M.current_tasks = tasks
  
  -- Apply filters first
  local filtered_tasks = apply_filters(tasks)
  
  -- Add filter indicators to title
  local filter_parts = {}
  if M.current_filters.status then
    table.insert(filter_parts, 'Status:' .. M.current_filters.status)
  end
  if #M.current_filters.tags > 0 then
    table.insert(filter_parts, 'Tags:' .. table.concat(M.current_filters.tags, ','))
  end
  if M.current_filters.deadline then
    table.insert(filter_parts, 'Deadline:' .. M.current_filters.deadline)
  end
  if M.current_filters.search then
    table.insert(filter_parts, 'Search:' .. M.current_filters.search)
  end
  
  if #filter_parts > 0 then
    title = title .. ' {' .. table.concat(filter_parts, ' ') .. '}'
  end
  
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
  local indicator_info = {}  -- { line_number = indicators }
  M.line_to_task_id = {}  -- Reset line to task ID mapping
  
  -- Build a map of parent tasks and their children using filtered tasks
  local parent_map = {}  -- parent_id -> list of child tasks
  local root_tasks = {}  -- tasks without parents
  local task_by_id = {}  -- quick lookup map
  
  for _, task in ipairs(filtered_tasks) do
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
    local task_lines, deadline_status, task_id, indicators = utils.format_task(task, M.current_view_mode)
    
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
    
    -- Store indicators for virtual text
    if indicators and #indicators > 0 then
      indicator_info[main_line_num] = indicators
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
  table.insert(lines, 'Keys: <CR> open  m* commands  ? help  q quit  ' .. mode_indicator)
  
  -- Set buffer content
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  -- Make buffer modifiable for direct editing
  vim.api.nvim_buf_set_option(buf, 'modifiable', true)
  
  -- Apply syntax highlights with delay to ensure stability
  vim.schedule(function()
    if vim.api.nvim_buf_is_valid(buf) then
      highlights.apply_highlights(buf)
    end
  end)
  
  -- Apply virtual text with delay to ensure buffer is ready
  vim.schedule(function()
    if vim.api.nvim_buf_is_valid(buf) then
      highlights.apply_deadline_virtual_text(buf, deadline_info)
      highlights.apply_task_id_virtual_text(buf, task_id_info, indicator_info)
      highlights.apply_indicator_virtual_text(buf, indicator_info)
    end
  end)
  
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
    
    -- Set up autocmd to close help window when main window closes
    vim.api.nvim_create_autocmd({'WinClosed', 'BufWipeout'}, {
      buffer = buf,
      callback = function()
        if M.help_win and vim.api.nvim_win_is_valid(M.help_win) then
          vim.api.nvim_win_close(M.help_win, true)
          M.help_win = nil
        end
      end,
      desc = 'Close help window when main mdtask window closes'
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
  
    -- ma to archive
    vim.keymap.set('n', 'ma', function()
      actions.quick_archive()
    end, opts)
    
    vim.keymap.set('n', 'mr', function()
      require('mdtask.tasks').list()
    end, opts)
    
    -- mn to create new task
    vim.keymap.set('n', 'mn', function()
      -- Don't close the window, just hide it temporarily
      require('mdtask.tasks').new()
    end, opts)
    
    -- me to edit task
    vim.keymap.set('n', 'me', function()
      local task_id = get_task_id_from_position()
      if task_id then
        require('mdtask.tasks').edit(task_id)
      end
    end, opts)
    
    -- ms to toggle status
    vim.keymap.set('n', 'ms', function()
      actions.toggle_task_status()
    end, opts)
    
    -- mp to preview
    vim.keymap.set('n', 'mp', function()
      actions.preview_task()
    end, opts)
    
    -- md to mark as DONE
    vim.keymap.set('n', 'md', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'DONE')
      end
    end, opts)
    
    -- mt to mark as TODO
    vim.keymap.set('n', 'mt', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'TODO')
      end
    end, opts)
    
    -- mw to mark as WIP
    vim.keymap.set('n', 'mw', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'WIP')
      end
    end, opts)
    
    -- mh to mark as SCHE (scheduled)
    vim.keymap.set('n', 'mh', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'SCHE')
      end
    end, opts)
    
    -- mz to mark as WAIT
    vim.keymap.set('n', 'mz', function()
      local task_id = get_task_id_from_position()
      if task_id then
        actions.quick_status_update(task_id, 'WAIT')
      end
    end, opts)
    
    -- Field-specific edits
    -- mS to edit status (with selection dialog)
    vim.keymap.set('n', 'mS', function()
      local task_id = get_task_id_from_position()
      if task_id then
        require('mdtask.tasks').edit_field(task_id, 'status')
      end
    end, opts)
    
    -- mT to edit title
    vim.keymap.set('n', 'mT', function()
      local task_id = get_task_id_from_position()
      if task_id then
        require('mdtask.tasks').edit_field(task_id, 'title')
      end
    end, opts)
    
    -- mD to edit description
    vim.keymap.set('n', 'mD', function()
      local task_id = get_task_id_from_position()
      if task_id then
        require('mdtask.tasks').edit_field(task_id, 'description')
      end
    end, opts)
    
    -- m/ to search tasks
    vim.keymap.set('n', 'm/', function()
      vim.ui.input({ prompt = 'Search tasks: ' }, function(query)
        if query and query ~= '' then
          vim.api.nvim_win_close(win, true)
          require('mdtask.tasks').search(query)
        end
      end)
    end, opts)
    
    -- mW to open web interface
    vim.keymap.set('n', 'mW', function()
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
    
    -- Filter commands
    -- f to show filter menu
    vim.keymap.set('n', 'f', function()
      local filter_options = {
        'status - Filter by status',
        'tags - Filter by tags',
        'deadline - Filter by deadline',
        'search - Search in tasks',
        'clear - Clear all filters',
      }
      
      vim.ui.select(filter_options, {
        prompt = 'Select filter type:',
        format_item = function(item)
          return item
        end,
      }, function(choice)
        if choice then
          local filter_type = choice:match('^(%S+)')
          
          if filter_type == 'status' then
            local status_options = {'TODO', 'WIP', 'WAIT', 'SCHE', 'DONE', 'clear'}
            vim.ui.select(status_options, {
              prompt = 'Filter by status:',
            }, function(status)
              if status == 'clear' then
                M.current_filters.status = nil
              elseif status then
                M.current_filters.status = status
              end
              if M.current_tasks then
                M.show_task_list(M.current_tasks)
              end
            end)
            
          elseif filter_type == 'tags' then
            vim.ui.input({ prompt = 'Filter by tags (comma-separated): ' }, function(tags_input)
              if tags_input == '' then
                M.current_filters.tags = {}
              elseif tags_input then
                M.current_filters.tags = {}
                for tag in tags_input:gmatch('[^,]+') do
                  table.insert(M.current_filters.tags, tag:match('^%s*(.-)%s*$'))
                end
              end
              if M.current_tasks then
                M.show_task_list(M.current_tasks)
              end
            end)
            
          elseif filter_type == 'deadline' then
            local deadline_options = {'overdue', 'today', 'week', 'none', 'clear'}
            vim.ui.select(deadline_options, {
              prompt = 'Filter by deadline:',
              format_item = function(item)
                local labels = {
                  overdue = 'overdue - Past due',
                  today = 'today - Due today',
                  week = 'week - Due this week',
                  none = 'none - No deadline',
                  clear = 'clear - Show all'
                }
                return labels[item] or item
              end,
            }, function(deadline)
              if deadline == 'clear' then
                M.current_filters.deadline = nil
              elseif deadline then
                M.current_filters.deadline = deadline
              end
              if M.current_tasks then
                M.show_task_list(M.current_tasks)
              end
            end)
            
          elseif filter_type == 'search' then
            vim.ui.input({ 
              prompt = 'Search in tasks: ',
              default = M.current_filters.search or ''
            }, function(search_term)
              if search_term == '' then
                M.current_filters.search = nil
              elseif search_term then
                M.current_filters.search = search_term
              end
              if M.current_tasks then
                M.show_task_list(M.current_tasks)
              end
            end)
            
          elseif filter_type == 'clear' then
            M.current_filters = {
              status = nil,
              tags = {},
              deadline = nil,
              search = nil
            }
            if M.current_tasks then
              M.show_task_list(M.current_tasks)
            end
          end
        end
      end)
    end, opts)
    
    -- Quick filter shortcuts
    vim.keymap.set('n', 'fs', function()
      -- Quick status filter toggle
      local statuses = {'TODO', 'WIP', 'WAIT', 'SCHE', 'DONE'}
      local current_index = 0
      for i, status in ipairs(statuses) do
        if M.current_filters.status == status then
          current_index = i
          break
        end
      end
      
      local next_index = (current_index % #statuses) + 1
      M.current_filters.status = statuses[next_index]
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    vim.keymap.set('n', 'fd', function()
      -- Quick deadline filter toggle
      local deadlines = {'overdue', 'today', 'week', 'none'}
      local current_index = 0
      for i, deadline in ipairs(deadlines) do
        if M.current_filters.deadline == deadline then
          current_index = i
          break
        end
      end
      
      local next_index = (current_index % #deadlines) + 1
      M.current_filters.deadline = deadlines[next_index]
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    vim.keymap.set('n', 'fc', function()
      -- Clear all filters
      M.current_filters = {
        status = nil,
        tags = {},
        deadline = nil,
        search = nil
      }
      if M.current_tasks then
        M.show_task_list(M.current_tasks)
      end
    end, opts)
    
    -- Quick search
    vim.keymap.set('n', '/', function()
      vim.ui.input({ 
        prompt = 'Search: ',
        default = M.current_filters.search or ''
      }, function(search_term)
        if search_term == '' then
          M.current_filters.search = nil
        elseif search_term then
          M.current_filters.search = search_term
        end
        if M.current_tasks then
          M.show_task_list(M.current_tasks)
        end
      end)
    end, opts)
    
    -- Task operations
    vim.keymap.set('n', 'my', function()
      require('mdtask.tasks').copy_task()
    end, opts)
    
    vim.keymap.set('n', 'mP', function()
      require('mdtask.tasks').paste_task()
    end, opts)
    
    vim.keymap.set('n', 'mdd', function()
      require('mdtask.tasks').delete_task()
    end, opts)
    
    -- mN to create subtask
    vim.keymap.set('n', 'mN', function()
      require('mdtask.tasks').new_subtask()
    end, opts)
    
    -- mL to list subtasks
    vim.keymap.set('n', 'mL', function()
      require('mdtask.tasks').list_subtasks()
    end, opts)
    
    -- mv to toggle view mode
    vim.keymap.set('n', 'mv', function()
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
    
    -- mi to show statistics
    vim.keymap.set('n', 'mi', function()
      M.show_stats(M.current_tasks)
    end, opts)
    
    -- ? to show help
    vim.keymap.set('n', '?', function()
      local help_text = [[
Task List Shortcuts:

Navigation & View:
  <CR>    Open task file
  mp      Preview task
  q       Quit
  mr      Refresh list
  m/      Search tasks
  mW      Open web interface
  mv      Toggle view (compact/detailed)
  mi      Show task statistics
  ?       Show this help

Task Management:
  mn      New task
  mN      New subtask (for current task)
  me      Edit task (full form)
  ma      Archive task
  my      Copy task (without ID)
  mP      Paste task (create new)
  mdd     Delete task (with confirmation)
  mL      List subtasks of current task

Quick Status Change:
  ms      Toggle status
  mt      Set TODO
  mw      Set WIP (Work in Progress)
  mz      Set WAIT
  mh      Set SCHE (Scheduled)
  md      Set DONE

Field-Specific Edit:
  mS      Edit status (with dialog)
  mT      Edit title
  mD      Edit description

Sorting:
  o       Sort menu
  oc      Sort by created date (toggle)
  ou      Sort by updated date (toggle)
  ot      Sort by title (toggle)
  os      Sort by status
  od      Sort by deadline
  oO      Reset to default order

Filtering:
  f       Filter menu
  fs      Quick status filter (cycle)
  fd      Quick deadline filter (cycle)
  fc      Clear all filters
  /       Quick search

Direct Editing:
  :w      Save changes (edit mode)
]]
      -- Create help buffer
      local help_buf = vim.api.nvim_create_buf(false, true)
      vim.api.nvim_buf_set_lines(help_buf, 0, -1, false, vim.split(help_text, '\n'))
      vim.api.nvim_buf_set_option(help_buf, 'modifiable', false)
      vim.api.nvim_buf_set_option(help_buf, 'buftype', 'nofile')
      vim.api.nvim_buf_set_option(help_buf, 'swapfile', false)
      vim.api.nvim_buf_set_option(help_buf, 'bufhidden', 'delete')
      
      -- Save current window
      local current_win = vim.api.nvim_get_current_win()
      
      -- Get current window dimensions and position
      local current_win_config = vim.api.nvim_win_get_config(current_win)
      local current_width = vim.api.nvim_win_get_width(current_win)
      local current_height = vim.api.nvim_win_get_height(current_win)
      
      -- Calculate help window dimensions and position
      local help_width = math.floor(current_width * 0.4)  -- 40% of current window width
      local help_height = current_height - 4  -- Leave some margin
      
      -- Position help window to the right of current window
      local help_row = current_win_config.row or 0
      local help_col = (current_win_config.col or 0) + current_width - help_width
      
      -- Create floating help window
      local help_win = vim.api.nvim_open_win(help_buf, true, {
        relative = current_win_config.relative or 'editor',
        width = help_width,
        height = help_height,
        row = help_row,
        col = help_col,
        border = 'rounded',
        style = 'minimal',
        zindex = 100,  -- Ensure it's on top
      })
      
      -- Store help window reference globally
      M.help_win = help_win
      
      -- Set window options
      vim.api.nvim_win_set_option(help_win, 'number', false)
      vim.api.nvim_win_set_option(help_win, 'relativenumber', false)
      vim.api.nvim_win_set_option(help_win, 'cursorline', true)
      vim.api.nvim_win_set_option(help_win, 'wrap', false)
      
      -- Function to close help window properly
      local function close_help()
        if vim.api.nvim_win_is_valid(help_win) then
          vim.api.nvim_win_close(help_win, true)
        end
        M.help_win = nil
        if vim.api.nvim_win_is_valid(current_win) then
          vim.api.nvim_set_current_win(current_win)
        end
      end
      
      -- Keymaps for closing help
      local help_opts = { buffer = help_buf, silent = true }
      vim.keymap.set('n', 'q', close_help, help_opts)
      vim.keymap.set('n', '<Esc>', close_help, help_opts)
      vim.keymap.set('n', '?', close_help, help_opts)
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
  
  -- Make preview window larger (90% of screen)
  local win_width = vim.api.nvim_get_option('columns')
  local win_height = vim.api.nvim_get_option('lines')
  local width = math.floor(win_width * 0.9)
  local height = math.floor(win_height * 0.85)
  
  local buf, win = utils.create_float_win({
    width = width,
    height = height,
  })
  
  -- Store task data for saving
  vim.api.nvim_buf_set_var(buf, 'mdtask_preview_task', task)
  
  -- Add help text at the top
  table.insert(lines, 1, '# Task Preview (:w to save, :q to close)')
  table.insert(lines, 2, '')
  
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(buf, 'modifiable', true)  -- Make editable
  vim.api.nvim_buf_set_option(buf, 'buftype', '')
  vim.api.nvim_buf_set_option(buf, 'bufhidden', 'wipe')
  vim.api.nvim_buf_set_option(buf, 'filetype', 'markdown')
  vim.api.nvim_buf_set_name(buf, 'mdtask-preview-' .. task.id)
  
  -- Save function
  local function save_preview()
    local content_lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
    local in_content = false
    local content_parts = {}
    local new_title = nil
    local new_description = nil
    local new_status = nil
    
    for i, line in ipairs(content_lines) do
      -- Skip help line
      if i == 1 and line:match('^# Task Preview') then
        goto continue
      end
      
      -- Extract title
      if not new_title and line:match('^# (.+)') then
        new_title = line:match('^# (.+)')
      -- Extract metadata
      elseif line:match('^%*%*Status:%*%* (.+)') then
        new_status = line:match('^%*%*Status:%*%* (.+)')
      elseif line:match('^%*%*Description:%*%* (.+)') then
        new_description = line:match('^%*%*Description:%*%* (.+)')
      -- Content starts after ---
      elseif line == '---' then
        in_content = true
      elseif in_content then
        table.insert(content_parts, line)
      end
      
      ::continue::
    end
    
    -- Remove trailing empty lines from content
    while #content_parts > 0 and content_parts[#content_parts] == '' do
      table.remove(content_parts)
    end
    
    local content = table.concat(content_parts, '\n')
    
    -- Update task if anything changed
    local args = {'edit', task.id}
    local has_changes = false
    
    if new_title and new_title ~= task.title then
      table.insert(args, '--title')
      table.insert(args, new_title)
      has_changes = true
    end
    
    if new_description and new_description ~= (task.description or '') then
      table.insert(args, '--description')
      table.insert(args, new_description)
      has_changes = true
    end
    
    if new_status and new_status ~= task.status then
      table.insert(args, '--status')
      table.insert(args, new_status)
      has_changes = true
    end
    
    if content ~= (task.content or '') then
      table.insert(args, '--content')
      table.insert(args, content)
      has_changes = true
    end
    
    if has_changes then
      utils.execute_mdtask(args, function(err, output)
        if err then
          utils.notify('Failed to save task: ' .. err, vim.log.levels.ERROR)
        else
          utils.notify('Task saved successfully')
          vim.api.nvim_buf_set_option(buf, 'modified', false)
          
          -- Refresh task list if it's open
          if M.task_list_buf and vim.api.nvim_buf_is_valid(M.task_list_buf) then
            require('mdtask.tasks').list()
          end
        end
      end)
    else
      utils.notify('No changes to save')
    end
  end
  
  -- Set up commands
  vim.api.nvim_buf_create_user_command(buf, 'w', save_preview, {})
  vim.api.nvim_buf_create_user_command(buf, 'wq', function()
    save_preview()
    vim.api.nvim_win_close(win, true)
  end, {})
  vim.api.nvim_buf_create_user_command(buf, 'q', function()
    if vim.api.nvim_buf_get_option(buf, 'modified') then
      utils.notify('No write since last change (add ! to override)', vim.log.levels.WARN)
    else
      vim.api.nvim_win_close(win, true)
    end
  end, {})
  vim.api.nvim_buf_create_user_command(buf, 'q!', function()
    vim.api.nvim_win_close(win, true)
  end, { bang = true })
  
  -- Set up keymaps
  local opts = { buffer = buf, silent = true }
  vim.keymap.set('n', '<C-s>', save_preview, opts)
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