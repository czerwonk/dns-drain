// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package gcloud

import (
	"google.golang.org/api/option"
)

type Config struct {
	Project         string
	CredentialsFile string
}

func (o Config) toClientOptions() []option.ClientOption {
	if o.CredentialsFile != "" {
		return []option.ClientOption{
			option.WithCredentialsFile(o.CredentialsFile),
		}
	}

	return []option.ClientOption{}
}
