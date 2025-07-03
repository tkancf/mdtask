local M = {}

-- Define highlight groups for mdtask
function M.setup()
  -- Status highlights
  vim.api.nvim_set_hl(0, 'MdTaskTodo', { link = 'Function', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskWip', { link = 'Type', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskWait', { link = 'Comment', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskSche', { link = 'Identifier', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskDone', { link = 'Comment', default = true })
  
  -- Section headers
  vim.api.nvim_set_hl(0, 'MdTaskHeader', { link = 'Title', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskSectionHeader', { bold = true, default = true })
  
  -- Other elements
  vim.api.nvim_set_hl(0, 'MdTaskDeadline', { link = 'WarningMsg', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskOverdue', { link = 'ErrorMsg', default = true })
  vim.api.nvim_set_hl(0, 'MdTaskLink', { link = 'Underlined', default = true })
end

-- Apply highlights to buffer
function M.apply_highlights(buf)
  -- Clear existing highlights
  vim.api.nvim_buf_clear_namespace(buf, -1, 0, -1)
  
  local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
  
  for i, line in ipairs(lines) do
    -- Highlight status brackets
    local status_match = line:match('^%s*%[(%w+)%]')
    if status_match then
      local start_col = line:find('%[')
      local end_col = line:find('%]') + 1
      
      if start_col then
        local hl_group = 'MdTaskTodo'
        if status_match == 'WIP' then
          hl_group = 'MdTaskWip'
        elseif status_match == 'WAIT' then
          hl_group = 'MdTaskWait'
        elseif status_match == 'SCHE' then
          hl_group = 'MdTaskSche'
        elseif status_match == 'DONE' then
          hl_group = 'MdTaskDone'
        end
        
        vim.api.nvim_buf_add_highlight(buf, -1, hl_group, i-1, start_col-1, end_col-1)
      end
    end
    
    -- Highlight section headers
    if line:match('^▸') then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskSectionHeader', i-1, 0, -1)
    end
    
    -- Highlight deadlines
    if line:match('%[OVERDUE%]') then
      local start_col = line:find('%[OVERDUE%]')
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskOverdue', i-1, start_col-1, start_col+8)
    elseif line:match('%[DUE%]') then
      local start_col = line:find('%[DUE%]')
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskDeadline', i-1, start_col-1, start_col+4)
    end
    
    -- Highlight markdown links
    local link_start, link_end = line:find('%[.-%]%(.-%)')
    if link_start then
      vim.api.nvim_buf_add_highlight(buf, -1, 'MdTaskLink', i-1, link_start-1, link_end)
    end
  end
end

return M