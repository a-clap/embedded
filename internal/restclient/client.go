/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package restclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func Get[T any, E error](url string, timeout time.Duration) (T, error) {
	ctx := context.Background()
	reqContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var get T
	// Prepare request
	r, err := http.NewRequestWithContext(reqContext, http.MethodGet, url, nil)
	if err != nil {
		return get, err
	}

	// Do
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return get, err
	}

	// If everything went fine, we should receive T
	if res.StatusCode == http.StatusOK {
		err = json.NewDecoder(res.Body).Decode(&get)
		return get, err
	}

	// Otherwise, error is expected
	var e E
	if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
		return get, err
	}
	return get, e
}

func Put[T any, E error](url string, timeout time.Duration, value T) (T, error) {
	ctx := context.Background()
	reqContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var put T
	b, err := json.Marshal(&value)
	if err != nil {
		return put, err
	}

	byteReader := bytes.NewReader(b)
	r, err := http.NewRequestWithContext(reqContext, http.MethodPut, url, byteReader)
	if err != nil {
		return put, err
	}

	// Header needed
	r.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return put, err
	}

	// If everything went fine, we should receive T
	if res.StatusCode == http.StatusOK {
		err = json.NewDecoder(res.Body).Decode(&put)
		return put, err
	}

	// Otherwise, error is expected
	var e E
	if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
		return put, err
	}
	return put, e
}
