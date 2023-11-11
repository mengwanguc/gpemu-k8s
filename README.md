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


### init master node

```
sudo kubeadm init --cri-socket --pod-network-cidr=10.244.0.0/16 unix:///var/run/cri-dockerd.sock 
```

You should see output like this:
```
Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

Alternatively, if you are the root user, you can run:

  export KUBECONFIG=/etc/kubernetes/admin.conf

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 10.52.2.142:6443 --token qnv6br.8xgtotg9sy6lu3pj \
        --discovery-token-ca-cert-hash sha256:f5cc26d25c380a445a688d11e9403762ce6f4cec355017e54fef038b0f06e916 
```

Try to run the commands as root user:

```
sudo -s
```

```
echo 'export KUBECONFIG=/etc/kubernetes/admin.conf' >> ~/.bashrc 
```

create a network pod

```
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
```

Check the network status

```
kubectl get pods --all-namespaces
```

reset
```

```