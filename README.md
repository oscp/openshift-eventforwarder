# General idea
We at [@SchweizerischeBundesbahnen](https://github.com/SchweizerischeBundesbahnen) need to host all the OpenShift events outside our OSE-cluster as it would flood our etcd datastore if we keep the events of all projects for more than one day.
So this tools just attaches to the kubernetes API and logs all the events to console where they are grabbed and sent to our central logging environment.

# Syslog forwarding
The standard mode for this applicaiton is to write to standard out.  If you wish to send to Syslog instead, you can define SYSLOG_SERVER as an environment variable, and we will forward the logs the the syslog server instead of sending the events to the console/standard out.  If you wish to send to both a syslog server as well as standard out, you can define the DEBUG environment variable and it will send to both standard out and the defined syslog server.

# Installation
```bash
# Create a project & a service-account
oc project logging
oc create serviceaccount ose-eventforwarder

# Add a new role to your cluster-policy:
oc create -f deploy/clusterPolicy-forward.yaml

# Add the role to the service-account
oc adm policy add-cluster-role-to-user ose:eventforwarder system:serviceaccount:logging:ose-eventforwarder

# Deploy the new pod
oc create configmap forward-config \
    --from-literal=syslog.server=\<syslogserver\>:\<syslogport\> \
    --from-literal=syslog.tag=\<syslog tag\>
oc create -f deploy/deploymentConfig.yaml
```

Just create a 'oc new-app' from building the dockerfile or get it from here [Dockerhub](https://hub.docker.com/r/oscp/openshift-eventforwarder/).

## Parameters
**Param**|**Description**|**Example**
:-----:|:-----:|:-----:
OPENSHIFT\_API\_URL|Your OpenShift API Url|https://master01.ch:8443
OPENSHIFT\_TOKEN|The token of the service-account| 
SYSLOG\_SERVER|The address and port of the target syslog server|syslogserver.corp.net:514
SYSLOG\_PROTO|Select tcp or udp for protocol. Defaults to udp if not defined| tcp
SYSLOG\_TAG|Tag to send to syslog identifying the source. Defaults to OSE if not defined| OSE\_CORP
DEBUG|Set to send to both standardout and syslog server. Defaults to FALSE | FALSE or TRUE
IGNORE\_SSL|Enable or disable SSL/TLS for the master api|Defaults to FALSE