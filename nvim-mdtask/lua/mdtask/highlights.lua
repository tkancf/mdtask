local M = {}

-- Define highlight groups for mdtask
function M.setup()
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
  
  -- UI elements
  vim.api.nvim_set_hl(0, 'MdTaskHeader', { fg = '#7aa2f7', bold = true, default = true })
  vim.api.nvim_set_hl(0, 'MdTaskSeparator', { fg = '#565f89', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskHelp', { fg = '#565f89', italic = true, default = true })
end

-- Apply highlights to buffer
function M.apply_highlights(buf)
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
        
        -- Check for deadline indicators
        local deadline_start = line:find('%[OVERDUE%]')
        local due_start = line:find('%[DUE%]')
        
        if deadline_start then
          vim.api.nvim_buf_add_highlight(buf, -1, title_hl, line_num, title_start, deadline_start - 1)
          vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskOverdue', line_num, deadline_start - 1, deadline_start + 8)
          if id_start and id_start > deadline_start + 8 then
            vim.api.nvim_buf_add_highlight(buf, -1, title_hl, line_num, deadline_start + 8, id_start - 1)
          end
        elseif due_start then
          vim.api.nvim_buf_add_highlight(buf, -1, title_hl, line_num, title_start, due_start - 1)
          vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskDeadline', line_num, due_start - 1, due_start + 4)
          if id_start and id_start > due_start + 4 then
            vim.api.nvim_buf_add_highlight(buf, -1, title_hl, line_num, due_start + 4, id_start - 1)
          end
        else
          vim.api.nvim_buf_add_highlight(buf, -1, title_hl, line_num, title_start, title_end)
        end
        
        -- Highlight task ID
        if id_start then
          vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskId', line_num, id_start - 1, -1)
        end
      end
    
    -- Description lines
    elseif line:match('^%s+%- ') and not line:match('^%s+%- %[.+%]%(') then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskDescription', line_num, 0, -1)
    
    -- Link lines
    elseif line:match('^%s+%- %[.+%]%(') then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskLink', line_num, 0, -1)
    end
  end
end

return M