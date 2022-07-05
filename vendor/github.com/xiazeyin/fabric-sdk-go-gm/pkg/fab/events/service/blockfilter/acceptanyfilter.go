/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockfilter

import (
	cb "github.com/xiazeyin/fabric-protos-go-gm/common"
	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/common/providers/fab"
)

// AcceptAny returns a block filter that accepts any block
var AcceptAny fab.BlockFilter = func(block *cb.Block) bool {
	return true
}
