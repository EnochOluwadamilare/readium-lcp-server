/*
 * Copyright (c) 2016-2018 Readium Foundation
 *
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 *  1. Redistributions of source code must retain the above copyright notice, this
 *     list of conditions and the following disclaimer.
 *  2. Redistributions in binary form must reproduce the above copyright notice,
 *     this list of conditions and the following disclaimer in the documentation and/or
 *     other materials provided with the distribution.
 *  3. Neither the name of the organization nor the names of its contributors may be
 *     used to endorse or promote products derived from this software without specific
 *     prior written permission
 *
 *  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
 *  ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 *  WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 *  DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
 *  ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 *  (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 *  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 *  ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 *  (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 *  SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package lutserver

import (
	"log"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/readium/readium-lcp-server/lib/http"
	"github.com/readium/readium-lcp-server/model"
)

// GetFilteredLicenses searches licenses activated by more than n devices
//
func GetFilteredLicenses(server http.IServer, req *http.Request) (model.LicensesCollection, error) {
	rDevices := req.FormValue("devices")
	log.Println("Licenses used by " + rDevices + " devices")
	if rDevices == "" {
		rDevices = "0"
	}

	if lic, err := server.Store().License().GetFiltered(rDevices); err == nil {
		return lic, nil
	} else {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, http.Problem{Detail: err.Error(), Status: http.StatusNotFound}
		default:
			return nil, http.Problem{Detail: err.Error(), Status: http.StatusInternalServerError}
		}
	}
}

// GetLicense gets an existing license by its id (passed as a section of the REST URL).
// It generates a partial license from the purchase info,
// fetches the license from the lcp server and returns it to the caller.
// This API method is called from a link in the license status document.
//
func GetLicense(server http.IServer, req *http.Request) (*model.License, error) {
	vars := mux.Vars(req)
	var licenseID = vars["license_id"]
	purchase, err := server.Store().Purchase().GetByLicenseID(licenseID)
	// get the license id in the URL
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, http.Problem{Detail: err.Error(), Status: http.StatusNotFound}
		default:
			return nil, http.Problem{Detail: err.Error(), Status: http.StatusInternalServerError}
		}
	}
	// get an existing license from the lcp server
	fullLicense, err := generateOrGetLicense(purchase, server)
	if err != nil {
		return nil, http.Problem{Detail: err.Error(), Status: http.StatusInternalServerError}

	}

	// message to the console
	log.Println("Get license / id " + licenseID + " / " + purchase.Publication.Title + " / purchase " + strconv.FormatInt(purchase.ID, 10))
	nonErr := http.Problem{Status: http.StatusOK, HttpHeaders: make(map[string][]string)}
	nonErr.HttpHeaders.Set(http.HdrContentDisposition, "attachment; filename=\"license.lcpl\"")
	return fullLicense, nonErr
}