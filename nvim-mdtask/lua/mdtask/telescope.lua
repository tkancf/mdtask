local M = {}

local has_telescope, telescope = pcall(require, 'telescope')
if not has_telescope then
  return M
end

local pickers = require('telescope.pickers')
local finders = require('telescope.finders')
local conf = require('telescope.config').values
local actions = require('telescope.actions')
local action_state = require('telescope.actions.state')
local previewers = require('telescope.previewers')

local utils = require('mdtask.utils')
local config = require('mdtask.config')

-- Telescope picker for tasks
function M.tasks(opts)
  opts = opts or {}
  local cfg = config.get()
  
  -- Get tasks
  local args = {'list', '--format', 'json'}
  if opts.status then
    table.insert(args, '--status')
    table.insert(args, opts.status)
  end
  if opts.all then
    table.insert(args, '--all')
  end
  
  utils.execute_mdtask(args, function(err, output)
    if err then
      utils.notify('Failed to get tasks: ' .. err, vim.log.levels.ERROR)
      return
    end
    
    local tasks = utils.parse_json(output)
    if #tasks == 0 then
      utils.notify('No tasks found')
      return
    end
    
    pickers.new(opts, {
      prompt_title = 'mdtask Tasks',
      finder = finders.new_table {
        results = tasks,
        entry_maker = function(task)
          return {
            value = task,
            display = utils.format_task(task),
            ordinal = task.title .. ' ' .. task.id .. ' ' .. (task.description or ''),
          }
        end
      },
      sorter = conf.generic_sorter(opts),
      previewer = cfg.telescope.show_preview and M.task_previewer() or nil,
      attach_mappings = function(prompt_bufnr, map)
        actions.select_default:replace(function()
          actions.close(prompt_bufnr)
          local selection = action_state.get_selected_entry()
          if selection then
            require('mdtask.tasks').edit(selection.value.id)
          end
        end)
        
        map('i', '<C-a>', function()
          local selection = action_state.get_selected_entry()
          if selection then
            require('mdtask.tasks').archive(selection.value.id)
            actions.close(prompt_bufnr)
          end
        end)
        
        map('n', 'a', function()
          local selection = action_state.get_selected_entry()
          if selection then
            require('mdtask.tasks').archive(selection.value.id)
            actions.close(prompt_bufnr)
          end
        end)
        
        return true
      end,
    }):find()
  end)
end

-- Task previewer
function M.task_previewer()
  return previewers.new_buffer_previewer {
    title = 'Task Preview',
    define_preview = function(self, entry, status)
      local task = entry.value
      local lines = {
        'ID: ' .. task.id,
        'Title: ' .. task.title,
        'Status: ' .. task.status,
        'Created: ' .. (task.created or ''),
        'Updated: ' .. (task.updated or ''),
        '',
      }
      
      if task.description then
        table.insert(lines, 'Description:')
        table.insert(lines, task.description)
        table.insert(lines, '')
      end
      
      if task.tags and #task.tags > 0 then
        table.insert(lines, 'Tags: ' .. table.concat(task.tags, ', '))
        table.insert(lines, '')
      end
      
      if task.content then
        table.insert(lines, 'Content:')
        table.insert(lines, '─────────────────────')
        local content_lines = vim.split(task.content, '\n')
        for _, line in ipairs(content_lines) do
          table.insert(lines, line)
        end
      end
      
      vim.api.nvim_buf_set_lines(self.state.bufnr, 0, -1, false, lines)
      vim.api.nvim_buf_set_option(self.state.bufnr, 'filetype', 'markdown')
    end
  }
end

-- Search tasks with telescope
function M.search()
  vim.ui.input({ prompt = 'Search tasks: ' }, function(query)
    if not query or query == '' then
      return
    end
    
    local args = {'search', query, '--format', 'json'}
    
    utils.execute_mdtask(args, function(err, output)
      if err then
        utils.notify('Failed to search tasks: ' .. err, vim.log.levels.ERROR)
        return
      end
      
      local tasks = utils.parse_json(output)
      if #tasks == 0 then
        utils.notify('No tasks found for: ' .. query)
        return
      end
      
      pickers.new({}, {
        prompt_title = 'Search Results: ' .. query,
        finder = finders.new_table {
          results = tasks,
          entry_maker = function(task)
            return {
              value = task,
              display = utils.format_task(task),
              ordinal = task.title .. ' ' .. task.id .. ' ' .. (task.description or ''),
            }
          end
        },
        sorter = conf.generic_sorter({}),
        previewer = M.task_previewer(),
        attach_mappings = function(prompt_bufnr, map)
          actions.select_default:replace(function()
            actions.close(prompt_bufnr)
            local selection = action_state.get_selected_entry()
            if selection then
              require('mdtask.tasks').edit(selection.value.id)
            end
          end)
          return true
        end,
      }):find()
    end)
  end)
end

-- Setup telescope integration
function M.setup()
  telescope.register_extension({
    setup = function(ext_config, config)
      -- Extension setup
    end,
    exports = {
      tasks = M.tasks,
      search = M.search,
    },
  })
end

return M