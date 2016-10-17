package magnet_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/magnet"

	"testing"
)

func init() {
	// don't clutter up stdout during tests
	magnet.SetOutput(ioutil.Discard)
}

func TestMagnet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Magnet Suite")
}
