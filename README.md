# S3 CSI driver for Kubernetes

### Overview
This is a repository for [S3 Object Storage](https://aws.amazon.com/cn/s3/) [CSI](https://kubernetes-csi.github.io/docs/) driver, csi plugin name: `s3.csi.k8s.io`. This driver requires an existing object storage server that supports the `S3` protocol. It supports dynamic configuration of persistent volumes through persistent volume claims by creating new `buckets` under the object storage server.

### Kubernetes Compatibility Matrix
|driver version  | supported k8s version | status |
|----------------|-----------------------|--------|
|master branch   | 1.21+                 | Alpha  |


### Install driver on a Kubernetes
 - [helm](./charts)
 - [kubectl](./docs/install-s3-csi-driver.md)

### Contribution, and support

You can reach the maintainers of this project at:

- [issues](https://github.com/keington/s3-csi-driver/issues/new)
- [mail](mailto:keington@outlook.com)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[Creative Commons 4.0]: https://git.k8s.io/website/LICENSE

### License
Unless otherwise noted, the content of this repository is licensed under the [Creative Commons 4.0] license, and code is licensed under the [Apache 2.0](LICENSE).# S3 CSI driver for Kubernetes