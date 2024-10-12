# Rover Controller

## Purpose
This controller allows employees to remotely manage lunar rovers with a central deployer. Using this tool, an authenticated employee can do things like send a rover to a remote location to dig mines or build habitats.

## Building
> Note that these are suboptimal build guides and referencing my personal directories.

To test the application locally:
* Build the golang with `go build`
* Run the controller with `sudo KUBECONFIG=/home/tokugero/repos/wrccdc/tartarus/kubeconfig ROVER_IMAGE=192.168.220.13:5000/rover:v1.0.0 ./tartarus.moon.mine -controlpanel`
    * App works best as sudo to use port 80, but not strictly necessary
    * KUBECONFIG is necessary to connect to the correct kubernetes cluster
    * ROVER_IMAGE tells the controller from where the rover image will come. Note that this is relative to the cluster
    * ./tartarus.moon.mine is the name of the binary
    * -controlpanel is the flag to trigger controlpanel
* Run the rover with `sudo ./tartarus.moon.mine -rover`

## Deploying
Alternatively you can run the container, you'll need to do this to deploy the service:
* Build the container with `docker build -t 192.168.220.13:5000/rover:<tag> .`
* Run the controller container with `docker run --rm -it -e KUBECONFIG=/home/tokugero/repos/wrccdc/tartarus/kubeconfig -e ROVER_IMAGE=192.168.220.13:5000/rover:v1.0.0 192.168.220.13:5000/rover -controlpanel`
* Run the rover container with `docker run --rm -it 192.168.220.13:5000/rover -rover`
