package knowndecor

import "github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/decorators"

const (
	Kibu                = "kibu"
	KibuProvider        = "kibu:provider"
	KibuWorkflow        = "kibu:workflow"
	KibuWorkflowUpdate  = "kibu:workflow:update"
	KibuWorkflowQuery   = "kibu:workflow:query"
	KibuWorkflowSignal  = "kibu:workflow:signal"
	KibuWorkflowExecute = "kibu:workflow:execute"
	KibuActivity        = "kibu:activity"
	KibuActivityMethod  = "kibu:activity:method"
	KibuService         = "kibu:service"
	KibuServiceMethod   = "kibu:service:method"
)

var (
	IsKibu               = decorators.HasPrefix(Kibu)
	IsKibuProvider       = decorators.HasPrefix(KibuProvider)
	IsKibuWorkflow       = decorators.HasPrefix(KibuWorkflow)
	IsKibuWorkflowUpdate = decorators.HasPrefix(KibuWorkflowUpdate)
	IsKibuWorkflowQuery  = decorators.HasPrefix(KibuWorkflowQuery)
	IsKibuWorkflowSignal = decorators.HasPrefix(KibuWorkflowSignal)
	IsKibuWorkflowExec   = decorators.HasPrefix(KibuWorkflowExecute)
	IsKibuActivity       = decorators.HasPrefix(KibuActivity)
	IsKibuActivityMethod = decorators.HasPrefix(KibuActivityMethod)
	IsKibuService        = decorators.HasPrefix(KibuService)
	IsKibuServiceMethod  = decorators.HasPrefix(KibuServiceMethod)
)
