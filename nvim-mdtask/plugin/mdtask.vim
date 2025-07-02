" mdtask.nvim - Neovim plugin for mdtask integration
" Author: tkancf
" License: MIT

if exists('g:loaded_mdtask')
  finish
endif
let g:loaded_mdtask = 1

" Commands
command! -nargs=0 MdTaskList lua require('mdtask').list()
command! -nargs=0 MdTaskNew lua require('mdtask').new()
command! -nargs=* MdTaskSearch lua require('mdtask').search(<q-args>)
command! -nargs=? MdTaskStatus lua require('mdtask.tasks').list_by_status(<q-args>)
command! -nargs=0 MdTaskWeb lua require('mdtask').open_web()
command! -nargs=? MdTaskEdit lua require('mdtask').edit(<q-args>)
command! -nargs=? MdTaskArchive lua require('mdtask').archive(<q-args>)

" Telescope commands (if available)
if exists(':Telescope')
  command! -nargs=0 TelescopeMdTaskTasks lua require('telescope').extensions.mdtask.tasks()
  command! -nargs=0 TelescopeMdTaskSearch lua require('telescope').extensions.mdtask.search()
endif