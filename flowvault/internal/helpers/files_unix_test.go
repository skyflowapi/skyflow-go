//go:build unix

package helpers

import (
	"context"
	"net/http"
	"path/filepath"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

var _ = Describe("Files, FIFOs", func() {
	It("refuses a FIFO without blocking, in the check and when sending", func() {
		fifo := filepath.Join(GinkgoT().TempDir(), "pipe")
		Expect(syscall.Mkfifo(fifo, 0o600)).To(Succeed())

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(checkReadable(fifo)).To(MatchError(notRegularFile))
			// A path that passed the check and was then swapped for a FIFO is refused too.
			status, msg := PutFile(context.Background(), http.DefaultClient, "http://127.0.0.1", ResolvedFile{Path: fifo})
			Expect(status).To(Equal(fvcommon.UploadStatusFailed))
			Expect(msg).To(ContainSubstring(notRegularFile))
		}()
		Eventually(done, 5*time.Second).Should(BeClosed())
	})
})
