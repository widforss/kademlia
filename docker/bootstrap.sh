#!/bin/bash

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

echo --- Bootstrap network -----------------------------------------------------
echo
for cide in ${cidip[@]}
do
eid=`echo $cide | cut -d\: -f1 | cut -c-8`
cmd="/usr/bin/tmux new-session -d './kademlia ${cip[0]}:9000'"
res=`docker exec $eid sh -c "${cmd}"`
echo $eid: $cmd
done
