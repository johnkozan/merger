// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gopkg.in/olivere/elastic.v3/backoff"
)

func Retry(attempts int, sleep time.Duration, callback func() error) (err error) {
	b := backoff.NewExponentialBackoff(sleep, 5*time.Second)
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(b.Next())

		zlog.Warn("retrying after error", zap.Error(err))
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
