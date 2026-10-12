package ai

import (
	"backend/utils"
)

func GenerateCompletion(pDebug *utils.HelperStruct, pSystemPrompt string, pUserPrompt string, pTemperature float32) (string, error) {
	pDebug.Log(utils.Statement, "GenerateCompletion(+)")
	defer pDebug.Log(utils.Statement, "GenerateCompletion(-)")

	lAns, _, _, lErr := GenerateCompletionWithFallback(pDebug, pSystemPrompt, pUserPrompt, pTemperature)
	if lErr != nil {
		return "", utils.ErrReturn(lErr)
	}

	return lAns, nil
}
