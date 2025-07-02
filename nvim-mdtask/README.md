# mdtask.nvim

A Neovim plugin for seamless integration with [mdtask](https://github.com/tkancf/mdtask), a task management tool that uses Markdown files as task tickets.

## Features

- 📋 List and browse tasks in floating windows
- ✨ Create new tasks with a simple form interface
- 🔍 Search tasks by content
- 📝 Edit tasks directly from Neovim
- 📦 Archive completed tasks
- 🌐 Launch mdtask web interface
- 🔭 Telescope integration for enhanced task browsing
- ⌨️ Intuitive keybindings and commands

## Requirements

- Neovim 0.7+
- [mdtask](https://github.com/tkancf/mdtask) binary installed and in PATH
- Optional: [telescope.nvim](https://github.com/nvim-telescope/telescope.nvim) for enhanced UI

## Installation

### Using [lazy.nvim](https://github.com/folke/lazy.nvim)

```lua
{
  'tkancf/mdtask',
  submodules = false,
  dir = vim.fn.stdpath('data') .. '/lazy/mdtask',
  dependencies = {
    'nvim-telescope/telescope.nvim', -- optional
  },
  config = function()
    -- Add the nvim-mdtask subdirectory to runtimepath
    vim.opt.rtp:append(vim.fn.stdpath('data') .. '/lazy/mdtask/nvim-mdtask')
    require('mdtask').setup({
      -- configuration options
    })
  end,
}
```

Or more simply, if you only want the Neovim plugin:

```lua
{
  dir = '~/path/to/mdtask/nvim-mdtask',  -- Adjust path to your mdtask clone
  name = 'mdtask.nvim',
  dependencies = {
    'nvim-telescope/telescope.nvim', -- optional
  },
  config = function()
    require('mdtask').setup({
      -- configuration options
    })
  end,
}
```

### Using [packer.nvim](https://github.com/wbthomason/packer.nvim)

```lua
use {
  'tkancf/mdtask',
  rtp = 'nvim-mdtask',
  requires = {
    'nvim-telescope/telescope.nvim', -- optional
  },
  config = function()
    require('mdtask').setup()
  end
}
```

## Configuration

```lua
require('mdtask').setup({
  -- Path to mdtask binary (default: 'mdtask')
  mdtask_path = 'mdtask',
  
  -- Default paths to search for tasks (default: {'.'})
  task_paths = { '.', '~/tasks' },
  
  -- Web server port (default: 7000)
  web_port = 7000,
  
  -- Open browser when starting web server (default: true)
  open_browser = true,
  
  -- Telescope configuration
  telescope = {
    enabled = true,
    theme = 'dropdown',
    show_preview = true,
  },
  
  -- UI configuration
  ui = {
    width = 80,
    height = 20,
    border = 'rounded',
  },
  
  -- Task creation defaults
  task_defaults = {
    status = 'TODO',
    tags = {},
  },
})
```

## Usage

### Commands

The plugin provides a unified `:MdTask` command with subcommands:

| Command | Description |
|---------|-------------|
| `:MdTask` | List all active tasks (default) |
| `:MdTask list [status]` | List tasks, optionally filtered by status |
| `:MdTask new` | Create a new task |
| `:MdTask search <query>` | Search tasks |
| `:MdTask edit <id>` | Edit a task |
| `:MdTask archive <id>` | Archive a task |
| `:MdTask web` | Start web interface |
| `:MdTask status <status>` | List tasks by status (TODO, WIP, WAIT, DONE) |
| `:MdTask toggle <id>` | Toggle task status |
| `:MdTask preview <id>` | Preview task details |
| `:MdTask help` | Show help message |

### Telescope Integration

If you have telescope.nvim installed, use Telescope with:

```vim
:Telescope mdtask tasks
:Telescope mdtask search
```

Or use Lua:

```lua
require('telescope').extensions.mdtask.tasks()
require('telescope').extensions.mdtask.search()
```

### Keybindings

In task list windows:

| Key | Action |
|-----|--------|
| `<CR>` | Edit selected task |
| `a` | Archive selected task |
| `n` | Create new task |
| `r` | Refresh task list |
| `q` | Close window |

In task form:

| Key | Action |
|-----|--------|
| `<C-s>` | Save task |
| `<C-c>` | Cancel |

### Task Form

When creating or editing tasks, you'll see a form like this:

```markdown
# Task Form

Title: Your task title here

Description: Optional description

Status: TODO

Tags: tag1, tag2, tag3

# Content

Optional additional content in markdown format

# Instructions
- Edit the fields above
- Press <C-s> to save
- Press <C-c> to cancel
```

## API

You can also use the plugin programmatically:

```lua
local mdtask = require('mdtask')

-- List tasks
mdtask.list()

-- Create new task
mdtask.new()

-- Search tasks
mdtask.search('query')

-- Edit task
mdtask.edit('task/20241228143000')

-- Archive task
mdtask.archive('task/20241228143000')

-- Open web interface
mdtask.open_web()
```

## Workflow Examples

### Quick Task Creation

```vim
:MdTask new
" Fill in the form and press <C-s>
```

### Browse and Edit Tasks

```vim
:MdTask
" Navigate with j/k, press <CR> to edit
```

### Search and Filter

```vim
:MdTask search important
:MdTask status WIP
```

### Toggle Task Status

```vim
:MdTask toggle task/20241228143000
" Or just press a key on the selected task in the list
```

### Telescope Integration

```vim
:Telescope mdtask tasks
" Use telescope's fuzzy finding to locate tasks
```

## Integration with mdtask

This plugin works seamlessly with the mdtask CLI tool. All operations performed through the plugin will be reflected in the mdtask web interface and vice versa.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see the [LICENSE](LICENSE) file for details.