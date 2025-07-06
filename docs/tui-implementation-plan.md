# TUI Implementation Plan for mdtask

## Overview

This document outlines the implementation plan for adding a Terminal User Interface (TUI) to mdtask. The TUI will provide an interactive, keyboard-driven interface for managing tasks directly in the terminal.

## Technology Selection

### TUI Library Options

After researching Go TUI libraries, the following options were considered:

1. **Bubble Tea** (github.com/charmbracelet/bubbletea)
   - Modern, actively maintained
   - Elm-inspired architecture
   - Great documentation and examples
   - Rich ecosystem (Bubbles components, Lipgloss for styling)
   - **Recommended choice**

2. **tview** (github.com/rivo/tview)
   - Mature, feature-rich
   - Traditional callback-based architecture
   - Good for complex layouts
   - Less modern approach

3. **termui** (github.com/gizak/termui)
   - Dashboard-focused
   - Less suitable for interactive applications

### Decision: Bubble Tea

We'll use Bubble Tea for the following reasons:
- Clean, maintainable architecture
- Active community and development
- Excellent styling capabilities with Lipgloss
- Built-in support for modern terminal features

## Feature Requirements

### Core Features

1. **Task List View**
   - Display tasks with status, ID, title, and tags
   - Color-coded status indicators
   - Sortable by: creation date, deadline, status, title
   - Filterable by: status, tags, search term

2. **Task Detail View**
   - Full markdown content display
   - Editable frontmatter fields
   - Status transitions
   - Tag management

3. **Task Creation**
   - Quick task creation with title
   - Optional initial tags and deadline
   - Auto-generated ID

4. **Task Editing**
   - In-TUI markdown editor (basic)
   - External editor integration (preferred)
   - Real-time preview option

5. **Search and Filter**
   - Full-text search
   - Tag-based filtering
   - Status filtering
   - Date range filtering

### Navigation and Controls

- **Keyboard-driven**: All actions accessible via keyboard
- **Vim-style keybindings**: Optional vim navigation (j/k, gg/G, etc.)
- **Help system**: Context-sensitive help (? key)
- **Command palette**: Quick action access (Ctrl+P style)

## UI Layout Design

### Main Screen Layout

```
┌─────────────────────────────────────────────────────────┐
│ mdtask TUI                              [F]ilter [H]elp │
├─────────────────────────────────────────────────────────┤
│ Status │ ID                  │ Title              │ Tags │
├────────┼────────────────────┼────────────────────┼──────┤
│ TODO   │ task/20240103120000 │ Implement TUI      │ dev  │
│ WIP    │ task/20240102150000 │ Fix bug #123       │ bug  │
│ WAIT   │ task/20240101090000 │ Review PR          │      │
├─────────────────────────────────────────────────────────┤
│ [n]ew [e]dit [v]iew [d]elete [s]tatus [q]uit          │
└─────────────────────────────────────────────────────────┘
```

### Task Detail View

```
┌─────────────────────────────────────────────────────────┐
│ Task: task/20240103120000               [Esc] Back     │
├─────────────────────────────────────────────────────────┤
│ Status: TODO                                            │
│ Tags: dev, feature, tui                                 │
│ Deadline: 2024-01-15                                   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ # Implement TUI for mdtask                              │
│                                                         │
│ ## Requirements                                          │
│ - Use Bubble Tea framework                              │
│ - Support all basic operations                          │
│                                                         │
├─────────────────────────────────────────────────────────┤
│ [e]dit [s]tatus [t]ags [d]eadline                     │
└─────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
1. Set up Bubble Tea project structure
2. Implement basic task list model
3. Create list view with navigation
4. Add task loading from filesystem

### Phase 2: Core Features (Week 2)
1. Implement task detail view
2. Add status transitions
3. Create new task functionality
4. Basic filtering by status

### Phase 3: Editing (Week 3)
1. External editor integration
2. In-TUI field editing
3. Tag management
4. Deadline setting

### Phase 4: Advanced Features (Week 4)
1. Search functionality
2. Advanced filtering
3. Sorting options
4. Configuration support

### Phase 5: Polish (Week 5)
1. Help system
2. Command palette
3. Performance optimization
4. Error handling improvements

## Technical Architecture

### Package Structure

```
cmd/
  mdtask/
    main.go          # Entry point with TUI subcommand
internal/
  tui/
    app.go           # Main TUI application
    models/
      task.go        # Task list model
      filter.go      # Filter/search model
    views/
      list.go        # Task list view
      detail.go      # Task detail view
      help.go        # Help view
    components/
      taskitem.go    # Task list item component
      statusbar.go   # Status bar component
      input.go       # Input field component
    keys/
      bindings.go    # Keyboard bindings
    styles/
      theme.go       # Color themes and styles
```

### Key Design Patterns

1. **Model-View-Update**: Bubble Tea's Elm architecture
2. **Component Composition**: Reusable UI components
3. **State Management**: Centralized app state
4. **Event Handling**: Message-based updates

## Integration Points

1. **CLI Integration**
   - Add `mdtask tui` subcommand
   - Pass configuration and paths

2. **Task Service**
   - Reuse existing task loading/saving logic
   - Maintain compatibility with CLI operations

3. **Configuration**
   - Support TUI-specific settings in TOML
   - Keybinding customization
   - Theme selection

## Testing Strategy

1. **Unit Tests**: Model and component logic
2. **Integration Tests**: TUI command execution
3. **Manual Testing**: Interactive UI testing
4. **Accessibility Testing**: Screen reader compatibility

## Success Criteria

1. All basic task operations available in TUI
2. Performance: <100ms response time for common operations
3. Works in common terminals (iTerm2, Terminal.app, tmux)
4. Graceful degradation for limited terminals
5. Comprehensive keyboard navigation

## Future Enhancements

1. **Multiple panes**: Split view for list and detail
2. **Bulk operations**: Multi-select for batch updates
3. **Themes**: User-customizable color schemes
4. **Plugins**: Extension system for custom views
5. **Sync indicators**: Real-time file system monitoring

## Implementation Status

### Completed Features ✓

1. **Basic TUI Infrastructure** ✓
   - Set up Bubble Tea dependencies
   - Created TUI subcommand
   - Implemented project structure

2. **Task List View** ✓
   - Display tasks with status, ID, and title
   - Keyboard navigation (j/k, arrow keys)
   - Search/filter functionality (built-in)
   - Help display

3. **Task Detail View** ✓
   - Full markdown content display
   - Task metadata display (ID, status, tags, deadline)
   - Navigation back to list

4. **Status Transitions** ✓
   - Status selector component
   - Keyboard shortcut 's' to change status
   - Persists changes to filesystem

5. **Multi-Select Functionality** ✓
   - Select/deselect individual tasks with 'v'
   - Select all tasks with 'a'
   - Clear all selections with 'A'
   - Bulk status changes for selected tasks
   - Visual indicator showing number of selected tasks

6. **Undo Functionality** ✓
   - Undo last status change with 'u' key
   - Tracks history of all status changes
   - Supports undoing bulk operations
   - Shows undo availability in status line
   - Groups related changes by timestamp

7. **Task Creation** ✓
   - Quick task creation with 'n' key
   - Interactive form with title, description, and tags
   - Tab navigation between fields
   - Ctrl+S to save, Esc to cancel
   - Auto-generates task ID and timestamps

8. **Task Editing** ✓
   - External editor integration with 'e' key
   - Configurable editor via config file or $EDITOR
   - Automatic reload after editing
   - Supports any text editor (vim, emacs, vscode, etc.)

### Remaining Features

3. **Advanced Features**
   - Tag management
   - Deadline setting
   - Archive/unarchive
   - Sorting options

## Usage

```bash
# Launch TUI
./mdtask tui

# Keyboard shortcuts
- j/k or ↑/↓: Navigate list
- Enter: View task details
- n: Create new task
- e: Edit task in external editor
- v: Select/deselect task for multi-select
- a: Select all tasks
- A: Clear all selections
- s: Change status (single task or bulk for selected tasks)
- u: Undo last status change
- /: Search tasks
- q: Quit
- ?: Show help

# Task creation form
- Tab/↓: Next field
- Shift+Tab/↑: Previous field
- Ctrl+S: Save task
- Esc: Cancel
```

## Next Steps

1. Implement task creation functionality
2. Add external editor integration
3. Add more keyboard shortcuts
4. Test with real task data