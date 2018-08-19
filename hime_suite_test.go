package hime_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHime(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hime Suite")
}
