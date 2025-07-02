" mdtask.nvim - Neovim plugin for mdtask integration
" Author: tkancf
" License: MIT

if exists('g:loaded_mdtask')
  finish
endif
let g:loaded_mdtask = 1

" Main command is defined in lua/mdtask/init.lua
" Old individual commands have been removed in favor of :MdTask subcommands

" Telescope integration is handled in lua/mdtask/init.lua