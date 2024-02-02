#!/bin/bash
# Usage: inside your ~/.tmux.conf set:
# set -g status-right '#[fg=colour0,bg=colour254 bold] 🔔 #( /path/to/todo/todo.tmux.sh ) | 🗓️   %d/%m #[fg=colour0,bg=colour254 bold]| ⏲️  %H:%M:%S '

DUE=$(todo due | wc -l)
ALL=$(todo list | wc -l)
echo "$ALL($DUE)"

