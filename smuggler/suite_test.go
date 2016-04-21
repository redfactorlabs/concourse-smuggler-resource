package smuggler_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
})

var _ = AfterSuite(func() {
})

func TestSmuggler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smuggler Suite")
}
