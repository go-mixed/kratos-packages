package rbac

import (
	"fmt"
	casbinLog "github.com/casbin/casbin/v2/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"strings"
)

type casbinLogger struct {
	logger *log.Helper
}

var _ casbinLog.Logger = &casbinLogger{}

func (l *casbinLogger) EnableLog(enable bool) {
}

func (l *casbinLogger) IsEnabled() bool {
	return true
}

func (l *casbinLogger) LogModel(model [][]string) {
	var str strings.Builder
	str.WriteString("Model: ")
	for _, v := range model {
		str.WriteString(fmt.Sprintf("%v\n", v))
	}

	l.logger.Debug(str.String())
}

func (l *casbinLogger) LogEnforce(matcher string, request []interface{}, result bool, explains [][]string) {
	var reqStr strings.Builder
	reqStr.WriteString("Request: ")
	for i, rval := range request {
		if i != len(request)-1 {
			reqStr.WriteString(fmt.Sprintf("%v, ", rval))
		} else {
			reqStr.WriteString(fmt.Sprintf("%v", rval))
		}
	}
	reqStr.WriteString(fmt.Sprintf(" ---> %t\n", result))

	reqStr.WriteString("Hit Policy: ")
	for i, pval := range explains {
		if i != len(explains)-1 {
			reqStr.WriteString(fmt.Sprintf("%v, ", pval))
		} else {
			reqStr.WriteString(fmt.Sprintf("%v \n", pval))
		}
	}

	l.logger.Debug(reqStr.String())
}

func (l *casbinLogger) LogPolicy(policy map[string][][]string) {
	var str strings.Builder
	str.WriteString("Policy: ")
	for k, v := range policy {
		str.WriteString(fmt.Sprintf("%s : %v\n", k, v))
	}

	l.logger.Debug(str.String())
}

func (l *casbinLogger) LogRole(roles []string) {
	l.logger.Debug("Roles: ", strings.Join(roles, "\n"))
}

func (l *casbinLogger) LogError(err error, msg ...string) {
	l.logger.Error(msg, err)
}
