# gpemu-k8s


## Set up SSH key

```
ssh-keygen -t rsa -b 4096
```

```
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
cat ~/.ssh/id_rsa.pub
```
Copy and paste into: https://github.com/settings/keys

### Configure github

NOTE: the example below is for Meng. You should modify it to your own account.

```
git config --global user.email "mengwanguc@gmail.com"
git config --global user.name "Meng Wang"
``

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


### Install kubeadm

```
bash setup/install-kubeadm.sh
```


### install cri-docker

```
cd /tmp
wget https://github.com/Mirantis/cri-dockerd/releases/download/v0.3.7/cri-dockerd_0.3.7.3-0.ubuntu-jammy_amd64.deb
sudo dpkg -i cri-dockerd_0.3.7.3-0.ubuntu-jammy_amd64.deb
cd -
```
