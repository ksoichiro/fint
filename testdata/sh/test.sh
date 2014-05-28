#!/bin/sh

# This is ok
something 2>&1

# This is NOT ok
something 2&>1

if [ -d ${HOME} ]; then echo "do something blur blur blur..."; else echo "do something"; fi
