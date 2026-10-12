package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	Elog      = 0
	Statement = 1
	Details   = 2
)

type ContextKey string

const RequestIDKey ContextKey = "requestID"

// InitFileLogger initializes file logging to a unique text file inside the "log" folder
func InitFileLogger() (*os.File, string, error) {
	lLogDir := "log"
	if lErr := os.MkdirAll(lLogDir, 0755); lErr != nil {
		return nil, "", lErr
	}

	lTimestamp := time.Now().Format("20060102_150405")
	lUID := strings.ReplaceAll(uuid.New().String(), "-", "")[:6]
	lFileName := fmt.Sprintf("log_%s_%s.txt", lTimestamp, lUID)
	lFilePath := filepath.Join(lLogDir, lFileName)

	lFile, lErr := os.OpenFile(lFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if lErr != nil {
		return nil, "", lErr
	}

	log.SetOutput(lFile)
	return lFile, lFilePath, nil
}

type HelperStruct struct {
	Sid       string
	Reference string
}

func (h *HelperStruct) Init() {
	h.Sid = strings.ReplaceAll(uuid.New().String(), "-", "")
}

func (h *HelperStruct) SetUid(r *http.Request) {
	requestID, ok := r.Context().Value(RequestIDKey).(string)
	if !ok {
		h.Init()
	} else {
		h.Sid = requestID
	}
}

func (h *HelperStruct) Log(pDebugLevel int, pMsg ...interface{}) {
	lPc, lFile, lLine, _ := runtime.Caller(1)
	lFuncname := runtime.FuncForPC(lPc).Name()

	// Hardcoded debug level for simplicity, normally read from config
	intlevel := 1 
	
	lStrArray := strings.Split(lFile, "/")
	lFilename := lFile
	if len(lStrArray) >= 2 {
		lFilename = lStrArray[len(lStrArray)-2] + "/" + lStrArray[len(lStrArray)-1]
	} else {
		lStrArray = strings.Split(lFile, "\\")
		if len(lStrArray) >= 2 {
			lFilename = lStrArray[len(lStrArray)-2] + "/" + lStrArray[len(lStrArray)-1]
		}
	}

	if (pDebugLevel <= intlevel && intlevel != 0) || pDebugLevel == Elog {
		str := fmt.Sprintf("%v", pMsg)
		lFinal := str[1 : len(str)-1]

		if h.Sid == "" {
			log.Println("Set the id before --> debug := new(helpers.HelperStruct); debug.Init()")
		} else {
			log.Println("@@", h.Sid, "@@ (", h.Reference, ") @@",
				lFilename, "@@", lFuncname, "@@ ln", lLine, "@@", lFinal)
		}
	}
}

type ErrorResponse struct {
	Status       string `json:"status"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"message"`
}

func GetErrorString(errCode string, errMsg string) string {
	resp := ErrorResponse{
		Status:       "E",
		ErrorCode:    errCode,
		ErrorMessage: errMsg,
	}
	bytes, _ := json.Marshal(resp)
	return string(bytes)
}

func ErrReturn(pErr error) error {
	if pErr == nil {
		return nil
	}
	lPc, lFile, lLine, _ := runtime.Caller(1)
	lFuncname := runtime.FuncForPC(lPc).Name()
	lStrArray := strings.Split(lFile, "/")
	lFilename := lFile
	if len(lStrArray) >= 2 {
		lFilename = lStrArray[len(lStrArray)-2] + "/" + lStrArray[len(lStrArray)-1]
	} else {
		lStrArray = strings.Split(lFile, "\\")
		if len(lStrArray) >= 2 {
			lFilename = lStrArray[len(lStrArray)-2] + "/" + lStrArray[len(lStrArray)-1]
		}
	}
	lErr := lFilename + " @@ " + lFuncname + " @@ ln " + strconv.Itoa(lLine) + " @@ "
	return fmt.Errorf("%s%w", lErr, pErr)
}
