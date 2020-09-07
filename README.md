# Kademlia

This is a Kademlia implementation written in the course D7024E at LTU in
fall 2020.

# Installation of Docker packages
Install Docker packages (exact command may be environment dependent).

apt-get install docker.io
apt-get install docker-compose

Note: You may need to use sudo during installation or use.

# Docker swarm

Specify desired number of replicated containers in  `Dockerfile` in line for `replicas`

    deploy:
      mode: replicated
      replicas: 50

Change directory to `docker`

    cd docker

Build the `kadlab` image

    docker build . -t kadlab

Initialize a swarm

    docker swarm init

Deploy the containers

    docker stack deploy --compose-file docker-compose.yml kadstack

To test communication between containers, We first find the container id by listing all active containers:

    docker ps

To determine the overlay network `kademlia_network` address of a container, inspect the container and look for `NetworkSettings.Networks.kademlia_kademlia_network.IPAddress`:

    docker inspect <CONTAINER_ID>
    (IP is given directly with: --format='{{<DATA PATH>}}')
   
We the container ID to execute a shell on the container

    docker exec -it <CONTAINER_ID> /bin/sh

From there, we can ping another container

    ping <IP_ADDR>

We can also use the [ping2host.sh](docker/ping2host.sh) script to see that any node can ping any other node.

    ./ping2host.sh

Notes for usage:
* `^P^Q` to deattach
* Tear down stack (crudely, hints that the containers should be nicely terminated): `docker stack rm kadstack`
* Inspect networks `docker network <ls / inspect <NETWORK_ID>>`

References:
* Docker documentation and man-pages.
