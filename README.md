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
```

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

https://linuxconfig.org/how-to-create-a-kubernetes-cluster

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


### init master node

```
sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --cri-socket unix:///var/run/cri-dockerd.sock --node-name scheduler
```

You should see output like this:
```
Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 10.52.2.142:6443 --token s3bpyl.5ad51wzm6o6jubyd \
        --discovery-token-ca-cert-hash sha256:7a46f5714c5b0dbf7024833ce21196842b54b2acc8f3a81057e39818e0b538d6
```

*Take a note of your own output , as you will need it to join the cluster*

Set up configuration file so that kubectl can run as non-root user.

```
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```


create a network pod

```
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
```

Check the network status

```
kubectl get pods --all-namespaces
```

### Run on worker node

Join the network

```
sudo kubeadm join 10.52.2.142:6443 --token s3bpyl.5ad51wzm6o6jubyd \
        --discovery-token-ca-cert-hash sha256:7a46f5714c5b0dbf7024833ce21196842b54b2acc8f3a81057e39818e0b538d6 \
        --cri-socket unix:///var/run/cri-dockerd.sock \
        --node-name worker-1
```

If you have multiple workers, make sure you modify the `node-name` above because Kubernetes doesn't
allow workers to have same name.




### Reset kubeadm
```
sudo kubeadm reset --cri-socket unix:///var/run/cri-dockerd.sock


sudo rm -r /etc/cni/net.d
sudo rm -rf ~/.kube /etc/kubernetes /var/lib/etcd /var/lib/kubelet

```