This folder provides reference files as well as samples for integrating Microcks into [Tekton](https://tekton.dev/) tasks and Pipelines.

### Reaching Microcks with Custom Certs

In case you have your Microcks installation between behind a TLS Ingress with custom certificate authority, you may have a look at the [`microcks-test-customcerts-task.yaml`](https://github.com/microcks/microcks-cli/blob/master/tekton/microcks-test-customcerts-task.yaml) that refer to an existing secret for retrieving this certificate.

You should have previously created your secret using something like this:

```sh
$ kubectl create secret generic microcks-test-customcerts-secret --from-file=ca.crt=ca.crt
```

And also having this secret, only accessible from a Service Account running your pipeline:

```sh
$ kubectl create -f ./microcks-test-customcerts-sa.yaml
```