package a

import (
	"context"
	"testing"
)

func cleanup(t *testing.T) {
	t.Context()
}

func f(ctx context.Context) {

}

func TestA(t *testing.T) {
	t.Cleanup(func() { t.Context() })          // want `avoid calling \(\*testing\.common\)\.Context inside Cleanup`
	t.Cleanup(func() { cleanup(t) })           // want `avoid calling \(\*testing\.common\)\.Context inside Cleanup`
	t.Cleanup(func() { f(t.Context()) })       // want `avoid calling \(\*testing\.common\)\.Context inside Cleanup`
	t.Cleanup(func() { context.Background() }) // ok
}
