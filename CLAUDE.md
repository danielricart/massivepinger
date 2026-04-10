# communication style
- be concise. 
- always plan before editing.
- be specific. 
- do not invent. in case of doubt, halt and ask questions. 
- try to parallelize.

# technical rules

## versions
- golang: 1.25.4
- dependencies: latest available at the moment of adding it

## dependencies
- optimize for native go libraries
- when using external dependencies, prioritize popular and standardized dependencies
  - logs: sirupsen/logrus
  - prometheus: prometheus-client
  - cli: spf13/cobra

## code structure
- code is modular
- filesystem organization from <project root>:
  - `main.go`
  - `/pkg` contains all modules
  - `/pkg/<module name>` contains each module's folder and their files within. all unit tests are in the same module folder with the suffix `_test` on each filename before the file extension (like: `metrics.go` `metrics_test.go`)
  - `helm-chart` contains the chart for kubernetes distribution. 

## build
building the project with native `go build` command

## testing 
testing the project with native `go test` command. Prioritize running always the entire test suite. this ensures that large code changes do not break other dependencies.
Each module has enough relevant tests and testcases to ensure a good coverage of each implemented feature.
End-to-end tests are also required to ensure that when the code is executed as intended, it behaves as defined (metrics are exposed, http endpoints respond, error loglines are generated, network requests to mocked third parties are established). These end-to-end tests are part of a dedicated module. end-to-end test files are also suffixed with  `_test`.

## distribution
### container
- the project distribution is a multi-stage docker file. 
  - one stage for dependency management. 
  - one for building. 
  - one for containering the binary. 
- everything based on official golang images matching the used golang version. 
- latest stage uses a distroless/static container. 
- the container has the relevant labels for metadata identification on GHCR.io.

### helm chart 
The project provides an easy path to deploy this exporter in a kubernetes cluster. The container requires the capability `NET_ADMIN` enabled or a more permissive environment. 

## CI
- github actions is used for Continuous Integration and building.
- on each PR that changes any *.go *.mod *.sum go tests is run.
- end to end tests are always run.
- on each PR that changes anything in `helm-chart` folder will run `helm lint` and `helm template` actions.
- on each merge to main:
  - all tests are run. 
  - the project is built. 
  - helm chart actions are run.
  - the container is built using `buildx` and for amd64 architecture. 
  - container is labelled as nightly. Only one nightly image is available. 
- On each github release creation:
  - all tests are run.
  - the project is built.
  - the container is built in a multiarch container for arm64 and amd64.
  - helm chart actions are run.
  - generated container is published to ghcr.io. 
  - container is labeled as `latest` and with the generated label from the release.

