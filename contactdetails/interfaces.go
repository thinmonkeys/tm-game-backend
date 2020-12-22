package contactdetails

type ContactDetailsProvider interface {
	GetContactDetails(cif string) (ContactDetails, error)
	SaveMobileNumber(cif string, newMobileNumber string) (err error)
	SaveHomeNumber(cif string, newHomeNumber string) (err error)
	SaveEmailAddress(cif string, newEmailAddress string) (err error)
}