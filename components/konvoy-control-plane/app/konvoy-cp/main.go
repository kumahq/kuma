package main

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client/example"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/postgres"
)

func main() {
	config := postgres.Config{
		Host: "localhost",
		Port: 5432,
		User: "postgres",
		Password: "mysecretpassword",
		DbName: "konvoy",
	}
	rc, err := postgres.NewStore(config)
	if err != nil {
		panic(err)
	}

	createResource := mesh.TrafficRouteResource{
		Spec: mesh.TrafficRoute{
			Path: "path-123",
		},
	}

	err = rc.Create(context.TODO(), &createResource, client.CreateByName("tr-1", "default"))
	if err != nil {
		panic(err)
	}

	resource := example.TrafficRouteResource{}
	err = rc.Get(context.TODO(), &resource, client.GetByName("tr-1", "default"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Get resource: %+v\n", &resource)

	resource.Spec.Path = "modified-path"
	err = rc.Update(context.TODO(), &resource)
	if err != nil {
		panic(err)
	}

	resourceAfterUpdate := example.TrafficRouteResource{}
	err = rc.Get(context.TODO(), &resourceAfterUpdate, client.GetByName("tr-1", "default"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Get resource after update: %+v\n", &resourceAfterUpdate)

	var trlist example.TrafficRouteResourceList
	err = rc.List(context.TODO(), &trlist, client.ListByNamespace("default"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("List of resources %+v\n", &trlist)

	err = rc.Delete(context.TODO(), &example.TrafficRouteResource{}, client.DeleteByName("tr-1", "default"))
	if err != nil {
		panic(err)
	}

	err = rc.List(context.TODO(), &trlist, client.ListByNamespace("default"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("List of resources %+v\n", &trlist)
}