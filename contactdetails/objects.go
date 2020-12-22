package contactdetails

type ContactDetails struct {
	CustomerCIF string
	FirstName string
	LastName string
	MobilePhoneNumber string
	HomePhoneNumber string
	EmailAddress string
	HomeAddress Address
}

type Address struct {
	FlatNumber string
	HouseName string
	HouseNumber string
	StreetName string
	District string
	Town string
	County string
	PostCode string
}

func BuildAddress(flat string, houseName string, houseNum string, street string, district string, town string, county string, postcode string) Address {
	return Address { flat, houseName, houseNum, street, district, town, county, postcode }
}

func BuildContactDetails(cif string, firstName string, lastName string, mobile string, home string, email string, address Address) ContactDetails {
	return ContactDetails{ cif, firstName, lastName, mobile, home, email, address }
}