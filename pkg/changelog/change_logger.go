// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package changelog

type ChangeLogger interface {
	LogChange(change DnsChange) error
}
