/*
 * Copyright (c) 2017, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/megaease/easegress/pkg/context"
	"github.com/megaease/easegress/pkg/logger"

	httpstat "github.com/tcnksm/go-httpstat"
)

type (
	request struct {
		server     *Server
		std        *http.Request
		statResult *httpstat.Result
		createTime time.Time
		_startTime *time.Time
		_endTime   *time.Time
	}

	resultState struct {
		buff *bytes.Buffer
	}
)

func (p *pool) newRequest(ctx context.HTTPContext, server *Server, reqBody io.Reader) (*request, error) {
	req := &request{
		createTime: time.Now(),
		server:     server,
		statResult: &httpstat.Result{},
	}

	r := ctx.Request()

	url := server.URL + r.Path()
	if r.Query() != "" {
		url += "?" + r.Query()
	}

	newCtx := httpstat.WithHTTPStat(ctx, req.statResult)
	stdr, err := http.NewRequestWithContext(newCtx, r.Method(), url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("BUG: new request failed: %v", err)
	}
	stdr.Header = r.Header().Std()

	req.std = stdr

	return req, nil
}

func (r *request) start() {
	if r._startTime != nil {
		logger.Errorf("BUG: started already")
		return
	}

	now := time.Now()
	r._startTime = &now
}

func (r *request) startTime() time.Time {
	if r._startTime == nil {
		return r.createTime
	}

	return *r._startTime
}

func (r *request) endTime() time.Time {
	if r._endTime == nil {
		return time.Now()
	}

	return *r._endTime
}

func (r *request) finish() {
	if r._endTime != nil {
		logger.Errorf("BUG: finished already")
		return
	}

	now := time.Now()
	r.statResult.End(now)
	r._endTime = &now
}

func (r *request) total() time.Duration {
	if r._endTime == nil {
		logger.Errorf("BUG: call total before finish")
		return r.statResult.Total(time.Now())
	}

	return r.statResult.Total(*r._endTime)
}

func (r *request) detail() string {
	rs := &resultState{buff: bytes.NewBuffer(nil)}
	r.statResult.Format(rs, 's')
	return rs.buff.String()
}

func (rs *resultState) Write(p []byte) (int, error) { return rs.buff.Write(p) }
func (rs *resultState) Width() (int, bool)          { return 0, false }
func (rs *resultState) Precision() (int, bool)      { return 0, false }
func (rs *resultState) Flag(c int) bool             { return false }
