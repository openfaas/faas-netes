// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"net/url"
	"strings"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

type FunctionLookup struct {
	DefaultNamespace string
}

func (l *FunctionLookup) Resolve(name string) (url.URL, error) {

	urlStr := fmt.Sprintf("http://%s.%s:%d", name, l.DefaultNamespace, watchdogPort)
	if strings.Contains(name, ".") {
		err := l.verifyNamespace(name)
		if err != nil {
			return url.URL{}, err
		}
		urlStr = fmt.Sprintf("http://%s:%d", name, watchdogPort)
	}

	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return url.URL{}, err
	}

	return *urlRes, nil
}

func (l *FunctionLookup) verifyNamespace(name string) error {
	// ToDo use global namepace parse and validation
	return nil
}
