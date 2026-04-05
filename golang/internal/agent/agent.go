package agent

import (
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
)

// ContractAgent 合同审查Agent接口
// Go版使用接口 + 结构体实现，对应Python的类和Java的Strategy模式。
type ContractAgent interface {
	Process(state *model.ReviewState) error
	Name() string
}
