package contactdetails

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	cd "../../contactdetails"
	"../common"
)


type ContactDetailsProvider struct {
	connection common.ConnectionSettings
}

func NewProvider() ContactDetailsProvider {
	return ContactDetailsProvider {
		connection: common.DefaultConnectionSettings(),
	}
}

type osCustomerInfo struct {
	CurrentAddress osAddressWrapper
	EmailAddress string
	MobilePhone string
	HomePhone string
	FirstName string
	LastName string
}

type osAddressWrapper struct {
	WOB_AddressInfo_IS osAddress
}
type osAddress struct {
	ApartmentNumber string
	HouseName string
	HouseNumber string
	Street string
	District string
	City string
	County string
	PostCode string
	Country string
	TimeAtAddressMonths int
	TimeAtAddressYears int
}

func (cdp ContactDetailsProvider) GetContactDetails(cif string) (cd.ContactDetails, error) {
	response, err := cdp.connection.RunRequest(http.MethodGet, fmt.Sprintf("/contactdetails/%s", cif), nil)
	if err != nil { return cd.ContactDetails{}, err }
	
	osDetails := osCustomerInfo{}
	err = json.NewDecoder(response.Body).Decode(&osDetails)
	if err != nil { return cd.ContactDetails{}, fmt.Errorf("Error decoding JSON response: %s", err.Error()) }

	osAddress := osDetails.CurrentAddress.WOB_AddressInfo_IS

	return cd.BuildContactDetails(
		cif, 
		osDetails.FirstName, 
		osDetails.LastName, 
		osDetails.MobilePhone, 
		osDetails.HomePhone, 
		osDetails.EmailAddress, 
		cd.BuildAddress(
			osAddress.ApartmentNumber, 
			osAddress.HouseName, 
			osAddress.HouseNumber, 
			osAddress.Street, 
			osAddress.District, 
			osAddress.City, 
			osAddress.County, 
			osAddress.PostCode)), nil
}

func (cdp ContactDetailsProvider) SaveMobileNumber(cif string, newMobileNumber string) (err error) {
	_,err = cdp.connection.RunRequest(http.MethodPut, fmt.Sprintf("/contactdetails/%s/mobile?MobileNumber=%s", cif, url.QueryEscape(newMobileNumber)), nil)
	return
}

func (cdp ContactDetailsProvider) SaveHomeNumber(cif string, newHomeNumber string) (err error) {
	_,err = cdp.connection.RunRequest(http.MethodPut, fmt.Sprintf("/contactdetails/%s/home?HomeNumber=%s", cif, url.QueryEscape(newHomeNumber)), nil)
	return
}

func (cdp ContactDetailsProvider) SaveEmailAddress(cif string, newEmailAddress string) (err error) {
	_,err = cdp.connection.RunRequest(http.MethodPut, fmt.Sprintf("/contactdetails/%s/email?EmailAddress=%s", cif, url.QueryEscape(newEmailAddress)), nil)
	return
}

func (cdp ContactDetailsProvider) SaveAddress(cif string, newAddress cd.Address) (err error) {
	newOsAddress := osAddress {
		ApartmentNumber: newAddress.FlatNumber,
		HouseName: newAddress.HouseName,
		HouseNumber: newAddress.HouseNumber,
		Street: newAddress.StreetName,
		District: newAddress.District,
		City: newAddress.Town,
		County: newAddress.County,
		PostCode: newAddress.PostCode,
		Country: "00826",
		TimeAtAddressMonths: 0,
		TimeAtAddressYears: 0,
	}
	_,err = cdp.connection.RunRequest(http.MethodPut, fmt.Sprintf("/contactdetails/%s/address", cif), newOsAddress)
	return
}


