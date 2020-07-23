module github.com/telexy324/grpc-resolver

go 1.14

require (
	github.com/etcd-io/etcd v3.3.22+incompatible // indirect
	github.com/hashicorp/consul v1.8.0 // indirect
	google.golang.org/grpc v1.30.0 // indirect
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
)

replace (
	labix.org/v2/mgo => github.com/go-mgo/mgo v0.0.0-20180705113738-7446a0344b78
	launchpad.net/gocheck => github.com/go-check/check v0.0.0-20200227125254-8fa46927fb4f
)
