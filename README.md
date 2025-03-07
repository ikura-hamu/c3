# c3

**C**leanup **C**ontext **C**hecker

c3 is a tool to analyze and warn against calling `(*testing.common).Context` within `(*testing.common).Cleanup` functions in Go tests.

`(*testing.common).Context` is canceled before `(*testing.common).Cleanup` functions are called, so calling `(*testing.common).Context` within `(*testing.common).Cleanup` functions can cause unexpected behavior.

> Context returns a context that is canceled just before Cleanup-registered functions are called.[^1]

[^1]: https://pkg.go.dev/testing#T.Context

```go
package a

import (
	"context"
	"testing"
)

func cleanup(t *testing.T) {
	t.Context()
}

func f(ctx context.Context) { }

func TestA(t *testing.T) {
	t.Cleanup(func() { t.Context() })          // want `avoid calling \(\*testing\.common\)\.Context inside Cleanup`
	t.Cleanup(func() { cleanup(t) })           // want `avoid calling \(\*testing\.common\)\.Context inside Cleanup`
	t.Cleanup(func() { f(t.Context()) })       // want `avoid calling \(\*testing\.common\)\.Context inside Cleanup`
	t.Cleanup(func() { context.Background() }) // ok
}
```

## Usage

### go vet

```bash
go install github.com/ikura-hamu/c3@latest
```

```bash
go vet -vettool=$(which c3) ./...
```

### golangci-lint Module Plugin

https://golangci-lint.run/plugins/module-plugins/

1. Configure golangci-lint customization.

`.custom-gcl.yml`

```yaml
version: "v1.64.1" # golangci-lint version
destination: "."
name: "custom-gcl"
plugins:
  - module: "github.com/ikura-hamu/c3"
    path: "github.com/ikura-hamu/c3"
    version: "latest"
```

2. Build custom golangci-lint bianry.

```bash
golangci-lint custom
```

3. Set up golangci-lint configuration.

`.golangci.yml`

```yml
linters:
  enable:
    - c3

linters-settings:
  custom:
    c3:
      type: "module"
```

4. Run custom golangci-lint.

```bash
./custom-gcl run
```
