kubectl run -l=type=mypod poddefault --image=praqma/network-multitool -- sleep 100000
kubectl run -l=type=mypod podfoo --serviceaccount=foo --image=praqma/network-multitool -- sleep 1000000
