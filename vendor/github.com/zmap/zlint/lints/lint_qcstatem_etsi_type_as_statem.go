/*
 * ZLint Copyright 2017 Regents of the University of Michigan
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
 * implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package lints

import (
	//"encoding/asn1"
	"github.com/zmap/zcrypto/encoding/asn1"
	"fmt"
	"github.com/zmap/zcrypto/x509"
	"github.com/zmap/zlint/util"
)

type qcStatemEtsiTypeAsStatem struct{}

func (l *qcStatemEtsiTypeAsStatem) Initialize() error {
	return nil
}

func (l *qcStatemEtsiTypeAsStatem) CheckApplies(c *x509.Certificate) bool {
	return util.IsExtInCert(c, util.QcStateOid)
}

func (l *qcStatemEtsiTypeAsStatem) Execute(c *x509.Certificate) *LintResult {
	errString := ""
	ext := util.GetExtFromCert(c, util.QcStateOid)

	oidList := make([]*asn1.ObjectIdentifier, 3)
	oidList[0] = &util.IdEtsiQcsQctEsign
	oidList[1] = &util.IdEtsiQcsQctEseal
	oidList[2] = &util.IdEtsiQcsQctWeb

	for _, oid := range oidList {
		r := util.ParseQcStatem(ext.Value, *oid)
		util.AppendToStringSemicolonDelim(&errString, r.GetErrorInfo())
		if r.IsPresent() {
			util.AppendToStringSemicolonDelim(&errString, fmt.Sprintf("ETSI QC Type OID %v used as QC statement", oid))
		}
	}

	if len(errString) == 0 {
		return &LintResult{Status: Pass}
	} else {
		return &LintResult{Status: Error, Details: errString}
	}
}

func init() {
	RegisterLint(&Lint{
		Name:          "e_qcstatem_etsi_type_as_statem",
		Description:   "Checks for erroneous QC Statement OID that actually are represented by ETSI ESI QC type OID.",
		Citation:      "ETSI EN 319 412 - 5 V2.2.1 (2017 - 11) / Section 4.2.3",
		Source:        EtsiEsi,
		EffectiveDate: util.EtsiEn319_412_5_V2_2_1_Date,
		Lint:          &qcStatemEtsiTypeAsStatem{},
	})
}
