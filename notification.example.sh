#!/bin/bash

if [ "$#" -ne 2 ]; then
	echo "Illegal number of parameters. Expected are exactly two (title and text) but are " $#
	exit 1
fi

notify-send -a "Todo" --hint='string:desktop-entry:todo' "$1" "$2"

