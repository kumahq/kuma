package v1alpha1_test

import (
	"fmt"

	discovery "github.com/Kong/konvoy/components/konvoy-control-plane-api/discovery/v1alpha1"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane-api/internal/util/proto"
)

func ExampleService() {
	bytes, _ := util_proto.ToYAML(&discovery.Service{
		Id: &discovery.Id{
			Namespace: "demo",
			Name:      "example",
		},
		Endpoints: []*discovery.Endpoint{
			&discovery.Endpoint{
				Address: "192.168.0.1",
				Port:    8080,
				Workload: &discovery.Workload{
					Id: &discovery.Id{
						Namespace: "demo",
						Name:      "example-xgg9df",
					},
					Meta: &discovery.Meta{
						Labels: map[string]string{
							"app":     "example",
							"version": "0.0.1",
						},
					},
					Locality: &discovery.Locality{
						Region: "us-east-1",
						Zone:   "us-east-1b",
					},
				},
			},
		},
	})
	fmt.Println(string(bytes))
	// Output:
	// endpoints:
	// - address: 192.168.0.1
	//   port: 8080
	//   workload:
	//     id:
	//       name: example-xgg9df
	//       namespace: demo
	//     locality:
	//       region: us-east-1
	//       zone: us-east-1b
	//     meta:
	//       labels:
	//         app: example
	//         version: 0.0.1
	// id:
	//   name: example
	//   namespace: demo
}
func ExampleWorkload() {
	bytes, _ := util_proto.ToYAML(&discovery.Workload{
		Id: &discovery.Id{
			Namespace: "demo",
			Name:      "daily-report-f5seg",
		},
		Meta: &discovery.Meta{
			Labels: map[string]string{
				"job": "daily-report",
			},
		},
		Locality: &discovery.Locality{
			Region: "us-west-2",
			Zone:   "us-west-2c",
		},
	})
	fmt.Println(string(bytes))
	// Output:
	// id:
	//   name: daily-report-f5seg
	//   namespace: demo
	// locality:
	//   region: us-west-2
	//   zone: us-west-2c
	// meta:
	//   labels:
	//     job: daily-report
}

func ExampleInventory() {
	bytes, _ := util_proto.ToYAML(&discovery.Inventory{
		Items: []*discovery.Inventory_Item{
			&discovery.Inventory_Item{
				ItemType: &discovery.Inventory_Item_Service{
					Service: &discovery.Service{
						Id: &discovery.Id{
							Namespace: "demo",
							Name:      "example",
						},
						Endpoints: []*discovery.Endpoint{
							&discovery.Endpoint{
								Address: "192.168.0.1",
								Port:    8080,
								Workload: &discovery.Workload{
									Id: &discovery.Id{
										Namespace: "demo",
										Name:      "example-xgg9df",
									},
									Meta: &discovery.Meta{
										Labels: map[string]string{
											"app":     "example",
											"version": "0.0.1",
										},
									},
									Locality: &discovery.Locality{
										Region: "us-east-1",
										Zone:   "us-east-1b",
									},
								},
							},
						},
					},
				},
			},
			&discovery.Inventory_Item{
				ItemType: &discovery.Inventory_Item_Workload{
					Workload: &discovery.Workload{
						Id: &discovery.Id{
							Namespace: "demo",
							Name:      "daily-report-f5seg",
						},
						Meta: &discovery.Meta{
							Labels: map[string]string{
								"job": "daily-report",
							},
						},
						Locality: &discovery.Locality{
							Region: "us-west-2",
							Zone:   "us-west-2c",
						},
					},
				},
			},
		},
	})
	fmt.Println(string(bytes))
	// Output:
	// items:
	// - service:
	//     endpoints:
	//     - address: 192.168.0.1
	//       port: 8080
	//       workload:
	//         id:
	//           name: example-xgg9df
	//           namespace: demo
	//         locality:
	//           region: us-east-1
	//           zone: us-east-1b
	//         meta:
	//           labels:
	//             app: example
	//             version: 0.0.1
	//     id:
	//       name: example
	//       namespace: demo
	// - workload:
	//     id:
	//       name: daily-report-f5seg
	//       namespace: demo
	//     locality:
	//       region: us-west-2
	//       zone: us-west-2c
	//     meta:
	//       labels:
	//         job: daily-report
}
