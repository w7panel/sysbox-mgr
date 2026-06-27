//
// Copyright 2026 Nestybox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package rootfsCloner

import (
	sh "github.com/nestybox/sysbox-libs/idShiftUtils"
	"github.com/nestybox/sysbox-mgr/lifecycleIO"
	"github.com/sirupsen/logrus"
)

func shiftIdsWithChown(path string, uidOffset, gidOffset int32) error {
	err := lifecycleIO.Default.Run(lifecycleIO.OperationChown, path, func() error {
		return sh.ShiftIdsWithChown(path, uidOffset, gidOffset)
	})
	logrus.Debugf("lifecycle IO stats: %s", lifecycleIO.Default.Stats())
	return err
}
