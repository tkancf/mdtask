local M = {}

-- Define highlight groups for mdtask
local function define_highlights()
  -- Status highlights
  vim.api.nvim_set_hl(0, 'MdTaskStatusTodo', { fg = '#7aa2f7', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskStatusWip', { fg = '#9ece6a', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskStatusWait', { fg = '#e0af68', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskStatusSche', { fg = '#bb9af7', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskStatusDone', { fg = '#565f89', default = true })
  
  -- Task elements
  vim.api.nvim_set_hl(0, 'MdTaskTitle', { link = 'Normal', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskTitleDone', { fg = '#565f89', strikethrough = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskId', { fg = '#565f89', italic = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskDescription', { fg = '#a9b1d6', italic = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskLink', { fg = '#7dcfff', underline = true, default = true })
  
  -- Deadline highlights
  vim.api.nvim_set_hl(0, 'MdTaskDeadline', { fg = '#e0af68', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskOverdue', { fg = '#f7768e', bold = true, default = true })
  
  -- Indicator highlights
  vim.api.nvim_set_hl(0, 'MdTaskIndicator', { fg = '#7dcfff', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskPriority', { fg = '#f7768e', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskToday', { fg = '#e0af68', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskSoon', { fg = '#ff9e64', bold = true, default = true })
  
  -- UI elements
  vim.api.nvim_set_hl(0, 'MdTaskHeader', { fg = '#7aa2f7', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskSeparator', { fg = '#565f89', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskHelp', { fg = '#565f89', italic = true, default = true })
end

function M.setup()
  -- Define highlights immediately
  define_highlights()
  
  -- Re-define highlights when colorscheme changes
  vim.api.nvim_create_autocmd('ColorScheme', {
    pattern = '*',
    callback = function()
      -- Use vim.schedule to ensure this runs after colorscheme is fully loaded
      vim.schedule(define_highlights)
    end,
    desc = 'Redefine mdtask highlight groups after colorscheme change'
  })
end

-- Apply highlights to buffer
function M.apply_highlights(buf)
  -- Ensure highlight groups are defined before applying
  -- This fixes issues when highlights aren't properly initialized
  local test_hl = vim.api.nvim_get_hl(0, { name = 'MdTaskStatusTodo' })
  if not test_hl or vim.tbl_isempty(test_hl) then
    define_highlights()
  end
  
  local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
  
  for i, line in ipairs(lines) do
    local line_num = i - 1  -- 0-indexed for nvim_buf_add_highlight
    
    -- Header (first line)
    if i == 1 then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskHeader', line_num, 0, -1)
    
    -- Separator lines
    elseif line:match('^[─═]+$') then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskSeparator', line_num, 0, -1)
    
    -- Help text (last line)
    elseif line:match('^Keys:') then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskHelp', line_num, 0, -1)
    
    -- Task lines
    elseif line:match('^%- ') then
      -- Extract status and apply appropriate highlight
      local status_start, status_end, status = line:find('^%- (%w+):')
      if status_start then
        local hl_group = 'MdTaskStatusTodo'
        local title_hl = 'MdTaskTitle'
        
        if status == 'WIP' then
          hl_group = 'MdTaskStatusWip'
        elseif status == 'WAIT' then
          hl_group = 'MdTaskStatusWait'
        elseif status == 'SCHE' then
          hl_group = 'MdTaskStatusSche'
        elseif status == 'DONE' then
          hl_group = 'MdTaskStatusDone'
          title_hl = 'MdTaskTitleDone'
        end
        
        -- Highlight status
        vim.api.nvim_buf_add_highlight(buf, -1, hl_group, line_num, 0, status_end)
        
        -- Highlight title
        local title_start = status_end + 1
        local id_start = line:find('{task/')
        local title_end = id_start and id_start - 1 or -1
        
        -- Highlight title
        vim.api.nvim_buf_add_highlight(buf, -1, title_hl, line_num, title_start, title_end)
        
        -- Highlight task ID
        if id_start then
          vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskId', line_num, id_start - 1, -1)
        end
      end
    
    -- Description lines
    elseif line:match('^%s+%- ') and not line:match('^%s+%- %[.+%]%(') then
      -- Check if it's a deadline line
      if line:match('^%s+%- Deadline:') then
        vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskDeadline', line_num, 0, -1)
      else
        vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskDescription', line_num, 0, -1)
      end
    
    -- Link lines
    elseif line:match('^%s+%- %[.+%]%(') then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskLink', line_num, 0, -1)
    end
  end
end

-- Apply virtual text for deadline indicators
function M.apply_deadline_virtual_text(buf, deadline_info)
  -- No longer needed since OVERDUE is shown as regular text
  -- Keep function for compatibility but do nothing
  local ns_id = vim.api.nvim_create_namespace('mdtask_deadline')
  vim.api.nvim_buf_clear_namespace(buf, ns_id, 0, -1)
end

-- Apply virtual text for task IDs with indicators
function M.apply_task_id_virtual_text(buf, task_id_info, indicator_info)
  -- Create namespace for virtual text
  local ns_id = vim.api.nvim_create_namespace('mdtask_id')
  
  -- Clear existing virtual text
  vim.api.nvim_buf_clear_namespace(buf, ns_id, 0, -1)
  
  -- Apply virtual text for each task ID, with indicators before the ID
  for line_num, task_id in pairs(task_id_info) do
    local virt_text = {}
    
    -- Add indicators first if they exist for this line
    local indicators = indicator_info[line_num]
    if indicators and #indicators > 0 then
      for _, indicator in ipairs(indicators) do
        local highlight = 'MdTaskIndicator'
        if indicator == '[!]' then
          highlight = 'MdTaskPriority'
        elseif indicator == '[OVERDUE]' then
          highlight = 'MdTaskOverdue'
        elseif indicator == '[TODAY]' then
          highlight = 'MdTaskToday'
        elseif indicator == '[SOON]' then
          highlight = 'MdTaskSoon'
        end
        table.insert(virt_text, {indicator .. ' ', highlight})
      end
    end
    
    -- Add task ID after indicators (with leading space)
    local id_prefix = (#virt_text > 0) and '' or ' '
    table.insert(virt_text, {id_prefix .. '{' .. task_id .. '}', 'MdTaskId'})
    
    vim.api.nvim_buf_set_extmark(buf, ns_id, line_num - 1, -1, {
      virt_text = virt_text,
      virt_text_pos = 'eol',
    })
  end
end

-- Apply virtual text for indicators (deprecated - now handled in apply_task_id_virtual_text)
function M.apply_indicator_virtual_text(buf, indicator_info)
  -- This function is now deprecated and does nothing
  -- Indicators are displayed with task IDs in apply_task_id_virtual_text
end

return M