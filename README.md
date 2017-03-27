# General idea
We at [@SchweizerischeBundesbahnen](https://github.com/SchweizerischeBundesbahnen) need to host all the OpenShift events outside our OSE-cluster as it would flood our etcd datastore if we keep the events of all projects for more than one day.
So this tools just attaches to the kubernetes API and logs all the events to console where they are grabbed and sent to our central logging environment.


# Installation
```bash
# Create a project & a service-account
oc new-project sbb-infra
oc create serviceaccount ose-eventforwarder

# Add a new role to your cluster-policy:
oc edit clusterPolicy default

###
- name: ose:eventforwarder
  role:
    metadata:
      creationTimestamp: null
      name: ose:eventforwarder
    rules:
    - apiGroups:
      - ""
      attributeRestrictions: null
      resources:
      - events
      verbs:
      - get
      - list
      - watch
###

# Add the role to the service-account
oc adm policy add-cluster-role-to-user ose:eventforwarder system:serviceaccount:sbb-infra:ose-eventforwarder
```

Just create a 'oc new-app' from building the dockerfile.

## Parameters
**Param**|**Description**|**Example**
:-----:|:-----:|:-----:
OPENSHIFT\_API\_URL|Your OpenShift API Url|https://master01.ch:8443
OPENSHIFT\_TOKEN|The token from the service-account| 

