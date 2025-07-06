local M = {}

-- Calculate task statistics
function M.calculate_stats(tasks)
  local stats = {
    total = 0,
    by_status = {
      TODO = 0,
      WIP = 0,
      WAIT = 0,
      SCHE = 0,
      DONE = 0,
    },
    overdue = 0,
    due_today = 0,
    due_this_week = 0,
    has_deadline = 0,
    no_deadline = 0,
    completion_rate = 0,
    created_today = 0,
    created_this_week = 0,
    created_this_month = 0,
    completed_today = 0,
    completed_this_week = 0,
    completed_this_month = 0,
  }
  
  -- Get current date info
  local now = os.time()
  local today = os.date("*t", now)
  today.hour = 0
  today.min = 0
  today.sec = 0
  local today_start = os.time(today)
  
  -- Calculate week start (Sunday)
  local days_since_sunday = today.wday - 1
  local week_start = today_start - (days_since_sunday * 24 * 60 * 60)
  
  -- Calculate month start
  local month_start = os.time({
    year = today.year,
    month = today.month,
    day = 1,
    hour = 0,
    min = 0,
    sec = 0
  })
  
  for _, task in ipairs(tasks) do
    stats.total = stats.total + 1
    
    -- Count by status
    local status = task.status or 'TODO'
    if stats.by_status[status] then
      stats.by_status[status] = stats.by_status[status] + 1
    end
    
    -- Check deadline
    if task.deadline then
      stats.has_deadline = stats.has_deadline + 1
      
      local deadline_time = vim.fn.strptime('%Y-%m-%dT%H:%M:%SZ', task.deadline)
      if deadline_time then
        if deadline_time < now then
          stats.overdue = stats.overdue + 1
        elseif deadline_time < today_start + (24 * 60 * 60) then
          stats.due_today = stats.due_today + 1
        elseif deadline_time < week_start + (7 * 24 * 60 * 60) then
          stats.due_this_week = stats.due_this_week + 1
        end
      end
    else
      stats.no_deadline = stats.no_deadline + 1
    end
    
    -- Check created date
    if task.created then
      local created_time = vim.fn.strptime('%Y-%m-%dT%H:%M:%SZ', task.created)
      if created_time then
        if created_time >= today_start then
          stats.created_today = stats.created_today + 1
        end
        if created_time >= week_start then
          stats.created_this_week = stats.created_this_week + 1
        end
        if created_time >= month_start then
          stats.created_this_month = stats.created_this_month + 1
        end
      end
    end
    
    -- Check completed date (if task is DONE)
    if status == 'DONE' and task.updated then
      local updated_time = vim.fn.strptime('%Y-%m-%dT%H:%M:%SZ', task.updated)
      if updated_time then
        if updated_time >= today_start then
          stats.completed_today = stats.completed_today + 1
        end
        if updated_time >= week_start then
          stats.completed_this_week = stats.completed_this_week + 1
        end
        if updated_time >= month_start then
          stats.completed_this_month = stats.completed_this_month + 1
        end
      end
    end
  end
  
  -- Calculate completion rate
  if stats.total > 0 then
    stats.completion_rate = math.floor((stats.by_status.DONE / stats.total) * 100)
  end
  
  return stats
end

-- Format stats for display
function M.format_stats(stats)
  local lines = {}
  
  -- Header
  table.insert(lines, "Task Statistics")
  table.insert(lines, string.rep("═", 40))
  table.insert(lines, "")
  
  -- Overall stats
  table.insert(lines, "Overall:")
  table.insert(lines, string.format("  Total Tasks: %d", stats.total))
  table.insert(lines, string.format("  Completion Rate: %d%%", stats.completion_rate))
  table.insert(lines, "")
  
  -- Status breakdown
  table.insert(lines, "By Status:")
  table.insert(lines, string.format("  TODO: %d", stats.by_status.TODO))
  table.insert(lines, string.format("  WIP:  %d", stats.by_status.WIP))
  table.insert(lines, string.format("  WAIT: %d", stats.by_status.WAIT))
  table.insert(lines, string.format("  SCHE: %d", stats.by_status.SCHE))
  table.insert(lines, string.format("  DONE: %d", stats.by_status.DONE))
  table.insert(lines, "")
  
  -- Deadline stats
  table.insert(lines, "Deadlines:")
  table.insert(lines, string.format("  Overdue: %d", stats.overdue))
  table.insert(lines, string.format("  Due Today: %d", stats.due_today))
  table.insert(lines, string.format("  Due This Week: %d", stats.due_this_week))
  table.insert(lines, string.format("  Has Deadline: %d", stats.has_deadline))
  table.insert(lines, string.format("  No Deadline: %d", stats.no_deadline))
  table.insert(lines, "")
  
  -- Activity stats
  table.insert(lines, "Activity:")
  table.insert(lines, string.format("  Created Today: %d", stats.created_today))
  table.insert(lines, string.format("  Created This Week: %d", stats.created_this_week))
  table.insert(lines, string.format("  Created This Month: %d", stats.created_this_month))
  table.insert(lines, string.format("  Completed Today: %d", stats.completed_today))
  table.insert(lines, string.format("  Completed This Week: %d", stats.completed_this_week))
  table.insert(lines, string.format("  Completed This Month: %d", stats.completed_this_month))
  table.insert(lines, "")
  
  -- Visual progress bar for completion rate
  table.insert(lines, "Progress:")
  local bar_width = 30
  local filled = math.floor((stats.completion_rate / 100) * bar_width)
  local empty = bar_width - filled
  local bar = "  [" .. string.rep("█", filled) .. string.rep("░", empty) .. "] " .. stats.completion_rate .. "%"
  table.insert(lines, bar)
  
  return lines
end

return M