package magnet_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMagnet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Magnet Suite")
}
