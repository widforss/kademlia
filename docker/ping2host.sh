#!/bin/bash

# References:
# Docker man pages and web-documentation
# Bash man page
# https://linuxhint.com/30_bash_script_examples

# Determine Container ID for all running Docker containers
# (--format could have been used but then it would have been necessary to remove header strings)
cids=`docker ps | grep kadstack_kademliaNodes | grep Up | cut -d\  -f1 | sort`

# Determine active container's Containter IDs

echo --- Detection phase ------
i=0
for cid in $cids
do
cidip[i]=`docker inspect --format='{{.ID}}:{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $cid`
cip[i]=`echo ${cidip[i]} | cut -d\: -f2`
echo ${cidip[i]}
i=$(($i+1))
done
echo $i running nodes.

echo

echo --- Ping test -----------------------------------------------------
echo Horizontally node IP addresses \(note\: a node may lack IP address\).
echo Vertically node id and ping exit status codes.
echo 0=reply and no error\; 1=no reply\; 2=other error.
echo
for cide in ${cidip[@]}
do
eid=`echo $cide | cut -d\: -f1 | cut -c-8`
res=`docker exec $eid sh ./ping2container.sh ${cip[*]}`
echo $eid:$res
done
