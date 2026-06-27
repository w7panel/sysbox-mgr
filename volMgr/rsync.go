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

package volMgr

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/nestybox/sysbox-libs/idShiftUtils"
	"github.com/nestybox/sysbox-mgr/lifecycleIO"
)

func (m *vmgr) rsyncVol(src, dest string, uid, gid uint32, shiftUids bool, shiftT shiftType) error {
	var output bytes.Buffer
	var usermap, groupmap string

	if shiftUids {
		srcUidList, srcGidList, err := idShiftUtils.GetDirIDs(src)
		if err != nil {
			return fmt.Errorf("failed to get user and group IDs for %s: %s", src, err)
		}

		usermap = rsyncIdMapOpt(srcUidList, uid, shiftT)
		groupmap = rsyncIdMapOpt(srcGidList, gid, shiftT)

		if usermap != "" {
			usermap = "--usermap=" + usermap
		}

		if groupmap != "" {
			groupmap = "--groupmap=" + groupmap
		}
	}

	srcDir := src + "/"
	args := []string{"-rauqlH", "--no-devices", "--delete"}
	if usermap != "" {
		args = append(args, usermap)
	}
	if groupmap != "" {
		args = append(args, groupmap)
	}
	args = append(args, srcDir, dest)

	cmd := exec.Command("rsync", args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := lifecycleIO.Default.Run(lifecycleIO.OperationRsync, dest, cmd.Run)
	if err != nil {
		return fmt.Errorf("rsync %s to %s: %v %v", srcDir, dest, string(output.Bytes()), err)
	}

	return nil
}

func rsyncIdMapOpt(idList []uint32, offset uint32, shiftT shiftType) string {
	var destId uint32

	mapOpt := ""
	for _, srcId := range idList {
		if shiftT == shiftUp {
			destId = srcId + offset
		} else {
			destId = srcId - offset
		}
		mapOpt += fmt.Sprintf("%d:%d,", srcId, destId)
	}

	if mapOpt != "" {
		mapOpt = strings.TrimSuffix(mapOpt, ",")
	}

	return mapOpt
}
