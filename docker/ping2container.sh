#!/bin/sh
for tip in "$@"
do
ping -W2 -c1 $tip > /dev/null 2>&1
echo -n $?\ 
done
echo
