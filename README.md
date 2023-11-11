# gpemu-k8s


## Set up SSH key

```
ssh-keygen -t rsa -b 4096
```

```
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
cat ~/.ssh/id_rsa.pub
```
Copy and paste into <a href="https://github.com/settings/keys" target="_blank">github ssh keys</a>

## clone this repo

```
git clone git@github.com:mengwanguc/gpemu-k8s.git
cd gpemu-k8s
```

## gitc

```
sudo cp gitc /usr/local/bin
```

## Install Docker

```
bash setup/install-docker.sh
```

### Verify docker installation

```
sudo docker run hello-world
```


### Install kubedam



