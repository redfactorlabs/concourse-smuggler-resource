package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var checkPath string
var inPath string
var outPath string

type suiteData struct {
	CheckPath string
	InPath    string
	OutPath   string
}

var _ = SynchronizedBeforeSuite(func() []byte {
	gp, err := gexec.Build("github.com/redfactorlabs/concourse-smuggler-resource/cmd")
	Ω(err).ShouldNot(HaveOccurred())
	gpDir := filepath.Dir(gp)
	cp := filepath.Join(gpDir, "check")
	os.Symlink(gp, cp)
	Ω(err).ShouldNot(HaveOccurred())
	ip := filepath.Join(gpDir, "in")
	os.Symlink(gp, ip)
	Ω(err).ShouldNot(HaveOccurred())
	op := filepath.Join(gpDir, "out")
	os.Symlink(gp, op)
	Ω(err).ShouldNot(HaveOccurred())

	data, err := json.Marshal(suiteData{
		CheckPath: cp,
		InPath:    ip,
		OutPath:   op,
	})
	Ω(err).ShouldNot(HaveOccurred())

	return data

}, func(data []byte) {
	var sd suiteData
	err := json.Unmarshal(data, &sd)
	Ω(err).ShouldNot(HaveOccurred())

	checkPath = sd.CheckPath
	inPath = sd.InPath
	outPath = sd.OutPath
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}
