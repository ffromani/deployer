/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2021 Red Hat, Inc.
 */

package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("[NegativeFlow] Deployer validation", func() {
	ginkgo.Context("with cluster with default settings", func() {
		ginkgo.It("it should fail the validation", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "deployer"),
				"validate",
				"--json",
			}
			fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = ginkgo.GinkgoWriter

			out, err := cmd.Output()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			vo := validationOutput{}
			if err := json.Unmarshal(out, &vo); err != nil {
				ginkgo.Fail(fmt.Sprintf("Error unmarshalling output %q: %v", out, err))
			}
			gomega.Expect(vo.Success).To(gomega.BeFalse())
			gomega.Expect(vo.Errors).ToNot(gomega.BeEmpty())
		})
	})
})

var _ = ginkgo.Describe("[NegativeFlow] Deployer option validation", func() {
	ginkgo.It("It should fail with invalid --updater-type", func() {
		updaterType := "LOCAL"
		err := deploy(updaterType)
		gomega.Expect(err).To(gomega.HaveOccurred(), "deployed succesfully with unknown updater type %s", updaterType)
	})
})
