# gpemu-k8s


# Set up SSH key

```
ssh-keygen -t rsa -b 4096
```

```
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
cat ~/.ssh/id_rsa.pub
```
Copy and paste into: https://github.com/settings/keys

## clone this repo

```
git clone git@github.com:mengwanguc/gpemu-k8s.git
cd gpemu-k8s
```

# gitc

cp gitc /usr/local/bin
