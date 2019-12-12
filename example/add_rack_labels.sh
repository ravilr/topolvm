#! /bin/sh

kubectl label nodes 192.168.1.101 rack=1
kubectl label nodes 192.168.1.102 rack=2
kubectl label nodes 192.168.1.103 rack=3

kubectl label nodes 192.168.1.101 zone=1
kubectl label nodes 192.168.1.102 zone=2
kubectl label nodes 192.168.1.103 zone=3

kubectl label nodes 192.168.1.101 failure-domain.beta.kubernetes.io/zone=1
kubectl label nodes 192.168.1.102 failure-domain.beta.kubernetes.io/zone=2
kubectl label nodes 192.168.1.103 failure-domain.beta.kubernetes.io/zone=3
