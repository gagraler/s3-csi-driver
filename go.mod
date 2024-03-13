module github.com/keington/s3-csi-driver

go 1.21.0

require (
	github.com/container-storage-interface/spec v1.9.0
	golang.org/x/net v0.22.0
	google.golang.org/grpc v1.62.1
	sigs.k8s.io/cloud-provider-azure v1.29.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/glog v1.2.0 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/rs/xid v1.5.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/moby/sys/mountinfo v0.7.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/client-go v0.29.2 // indirect
	k8s.io/klog/v2 v2.120.1
	k8s.io/mount-utils v0.29.2
	k8s.io/utils v0.0.0-20240310230437-4693a0247e57
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.17.0
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/minio/minio-go/v7 v7.0.69
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240311173647-c811ad7063a7 // indirect
	google.golang.org/protobuf v1.33.0
	k8s.io/apimachinery v0.29.2 // indirect
)

replace (
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.29.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.29.2

)
