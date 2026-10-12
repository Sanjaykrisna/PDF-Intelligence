package services

import (
	"fmt"
	"time"

	"backend/repositories"
	"backend/utils"
)

// getNextScheduledFlushTime calculates the next occurrence among 09:00, 12:00, 15:00, and 21:00.
func getNextScheduledFlushTime(pNow time.Time) time.Time {
	// 4 scheduled times per day: 9:00 AM, 12:00 PM, 3:00 PM, and 9:00 PM
	lTargetHours := []int{9, 12, 15, 21}
	for _, lHour := range lTargetHours {
		lTarget := time.Date(pNow.Year(), pNow.Month(), pNow.Day(), lHour, 0, 0, 0, pNow.Location())
		if lTarget.After(pNow) {
			return lTarget
		}
	}
	// If all 4 slots today have already passed, schedule for 09:00 AM tomorrow
	return time.Date(pNow.Year(), pNow.Month(), pNow.Day()+1, 9, 0, 0, 0, pNow.Location())
}

// StartDailyFlushScheduler runs 4 times every day: morning 9:00AM, 12:00PM, 3:00PM, and 9:00PM.
func StartDailyFlushScheduler(pDebug *utils.HelperStruct) {
	pDebug.Log(utils.Statement, "StartDailyFlushScheduler(+)")
	defer pDebug.Log(utils.Statement, "StartDailyFlushScheduler(-)")

	go func() {
		for {
			lNow := time.Now()
			lNextRun := getNextScheduledFlushTime(lNow)
			lDuration := lNextRun.Sub(lNow)

			lLogDebug := new(utils.HelperStruct)
			lLogDebug.Init()
			lLogDebug.Log(
				utils.Statement,
				"Daily 4x flush scheduler armed. Next execution at:",
				lNextRun.Format("2006-01-02 15:04:05"),
				fmt.Sprintf("(in %s)", lDuration.Round(time.Second).String()),
			)

			time.Sleep(lDuration)

			lExecDebug := new(utils.HelperStruct)
			lExecDebug.Init()
			lExecDebug.Log(utils.Statement, "ScheduledFlushTriggered(+)", time.Now().Format("2006-01-02 15:04:05"))

			lFlushedCount, lErr := repositories.FlushAllDocumentsAndChunks(lExecDebug)
			if lErr != nil {
				lExecDebug.Log(utils.Elog, "SCHED_FLUSH_01", "Error flushing documents during scheduled run:", lErr.Error())
			} else {
				lExecDebug.Log(utils.Statement, "Scheduled flush completed successfully. Total documents purged:", lFlushedCount)
			}
			lExecDebug.Log(utils.Statement, "ScheduledFlushTriggered(-)")

			// 2-minute safety delay so we do not re-trigger within the same minute
			time.Sleep(time.Minute * 2)
		}
	}()
}

// StartMidnightFlushScheduler is maintained for backward compatibility.
func StartMidnightFlushScheduler(pDebug *utils.HelperStruct) {
	StartDailyFlushScheduler(pDebug)
}

// TriggerManualFlush allows programmatic or test flushing of all stored documents and chunks.
func TriggerManualFlush(pDebug *utils.HelperStruct) (int64, error) {
	pDebug.Log(utils.Statement, "TriggerManualFlush(+)")
	defer pDebug.Log(utils.Statement, "TriggerManualFlush(-)")

	return repositories.FlushAllDocumentsAndChunks(pDebug)
}

