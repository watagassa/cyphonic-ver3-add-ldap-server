package entity_test

import (
	"github.com/Pluslab/cyphonic/dsd/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommonKey", func() {
	Context("GenerateCommonKey", func() {
		It("should generate a common key", func() {
			commonKey, err := entity.GenerateCommonKey()
			Expect(err).To(BeNil())
			Expect(len(commonKey)).To(Equal(32))
		})
	})
})
