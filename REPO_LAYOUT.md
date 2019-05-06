# Repository layout

* [components/](components/): Source code of `Konvoy`
* [distributions/](components/): Build scripts to generate deb/rpm/tar packages and Docker images
 with `Konvoy`
* [plugins/](plugins/): `Konvoy` extensions  

## [components/](components/)

* [components/konvoy-binary](components/konvoy-binary/): Build scripts to assemble Envoy binary 
 that includes 3rd party extensions, such as `Konvoy`, Istio, etc 
* [components/konvoy-filter](components/konvoy-filter/): Source code of `Konvoy`'s extensions 
 to Envoy

## [distributions/](distributions/)

* [distributions/Makefile](distributions/Makefile): Build script 
* [distributions/configs/](distributions/configs/): Sample Envoy configurations to include into 
 distributions
* [distributions/images/](distributions/images/): Docker images used by the build process
  * [distributions/images/fpm](distributions/images/fpm/): Docker image with FPM (build tool)
  * [distributions/images/konvoy](distributions/images/konvoy/): Docker images with `Konvoy binary`

## [plugins/](plugins/)

* [plugins/demo-grpc-server-java](plugins/demo-grpc-server-java/): TODO:
* [plugins/demo-grpc-server-go](plugins/demo-grpc-server-go/): TODO:
* [plugins/kafka-connector](plugins/kafka-connector/): TODO:
