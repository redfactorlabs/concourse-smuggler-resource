package smuggler_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var checkPath string

var _ = BeforeSuite(func() {
	var err error

	commands := []string{"check", "in", "out"}

	for _, c := range commands {
		checkPath, err = gexec.Build(fmt.Sprintf("github.com/redfactorlabs/concourse-smuggler-resource/cmd/%s", c))
		Ω(err).ShouldNot(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestSmuggler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smuggler Suite")
}

func Fixture(filename string) string {
	path := filepath.Join("fixtures", filename)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(contents)
}
