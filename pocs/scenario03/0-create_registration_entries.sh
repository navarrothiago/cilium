#/bin/bash


# agent
#kubectl exec -n spire spire-server-0 -- \
#    /opt/spire/bin/spire-server entry create \
#    -node  \
#    -spiffeID spiffe://example.org/ns/spire/sa/spire-agent \
#    -selector k8s_sat:cluster:demo-cluster \
#    -selector k8s_sat:agent_ns:spire \
#    -selector k8s_sat:agent_sa:spire-agent

kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://example.org/xwing \
    -parentID spiffe://example.org/ns/spire/sa/spire-agent \
    -selector k8s:pod-label:class:xwing 

kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://example.org/deathstar \
    -parentID spiffe://example.org/ns/spire/sa/spire-agent \
    -selector k8s:pod-label:class:deathstar 

